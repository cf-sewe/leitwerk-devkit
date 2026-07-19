package gate

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeChecks writes a set of scripts with given exit codes into a temp dir and
// returns a Resolver pointing at it as the built-in checks dir.
func fakeChecks(t *testing.T) Resolver {
	t.Helper()
	dir := t.TempDir()
	scripts := map[string]string{
		// exit 0, two lines of output — only the last must be shown.
		"pass.sh": "#!/bin/sh\necho first line\necho ok done\nexit 0\n",
		"skip.sh": "#!/bin/sh\necho nothing to run\nexit 2\n",
		"fail.sh": "#!/bin/sh\necho boom\nexit 1\n",
	}
	for name, body := range scripts {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return Resolver{LocalDir: filepath.Join(dir, "does-not-exist"), BuiltinDir: dir}
}

func runVerify(t *testing.T, tier, tiersBody string, r Resolver) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	code := RunVerify(VerifyOptions{
		Tier:      tier,
		Tiers:     ParseTiers([]byte(tiersBody)),
		Resolver:  r,
		TiersPath: "test-tiers",
		Stdout:    &out,
		Stderr:    &errb,
		Color:     false,
	})
	return code, out.String(), errb.String()
}

const verifyTiers = `[tiers]
TP = pass
TS = skip
TF = fail
TX = pass skip
TM = missing
`

func TestVerifyPass(t *testing.T) {
	code, out, _ := runVerify(t, "TP", verifyTiers, fakeChecks(t))
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if !strings.Contains(out, "gate: PASS") || !strings.Contains(out, "✓") || !strings.Contains(out, "ok done") {
		t.Errorf("unexpected output:\n%s", out)
	}
	if strings.Contains(out, "first line") {
		t.Errorf("only the last output line should be shown, got:\n%s", out)
	}
}

func TestVerifySkipDoesNotFail(t *testing.T) {
	code, out, _ := runVerify(t, "TS", verifyTiers, fakeChecks(t))
	if code != 0 {
		t.Errorf("exit = %d, want 0 (skip is not a failure)", code)
	}
	if !strings.Contains(out, "(skipped)") || !strings.Contains(out, "gate: PASS") {
		t.Errorf("unexpected output:\n%s", out)
	}
}

func TestVerifyFail(t *testing.T) {
	code, out, _ := runVerify(t, "TF", verifyTiers, fakeChecks(t))
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(out, "gate: FAIL") || !strings.Contains(out, "✗") || !strings.Contains(out, "boom") {
		t.Errorf("unexpected output:\n%s", out)
	}
}

func TestVerifyMixedSkipAndPass(t *testing.T) {
	code, out, _ := runVerify(t, "TX", verifyTiers, fakeChecks(t))
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if !strings.Contains(out, "gate: PASS") {
		t.Errorf("want PASS with a pass+skip mix:\n%s", out)
	}
}

func TestVerifyNoSuchCheck(t *testing.T) {
	code, out, _ := runVerify(t, "TM", verifyTiers, fakeChecks(t))
	if code != 1 {
		t.Errorf("exit = %d, want 1 (missing check fails the gate)", code)
	}
	if !strings.Contains(out, "no such check") || !strings.Contains(out, "?") {
		t.Errorf("unexpected output:\n%s", out)
	}
}

func TestVerifyUnknownTierIsUsageError(t *testing.T) {
	code, _, errb := runVerify(t, "T9", verifyTiers, fakeChecks(t))
	if code != 2 {
		t.Errorf("exit = %d, want 2 (no checks defined)", code)
	}
	if !strings.Contains(errb, "no checks defined for tier 'T9'") {
		t.Errorf("unexpected stderr: %q", errb)
	}
}

func TestVerifyColorProducesAnsi(t *testing.T) {
	var out, errb bytes.Buffer
	RunVerify(VerifyOptions{
		Tier: "TP", Tiers: ParseTiers([]byte(verifyTiers)), Resolver: fakeChecks(t),
		TiersPath: "x", Stdout: &out, Stderr: &errb, Color: true,
	})
	if !strings.Contains(out.String(), "\x1b[32m") {
		t.Errorf("expected ANSI colour codes when Color=true")
	}
}
