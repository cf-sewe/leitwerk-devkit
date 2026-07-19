package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Two copies of the freshly-built binary:
//
//	installBin — laid out with sibling checks/, templates/, leitwerk.tiers (a full
//	  checkout / release tarball): exercises on-disk resolution.
//	embedBin — the binary alone, no siblings (a `go install`): exercises the
//	  embedded-assets fallback, proving the gate does not depend on repo layout.
var (
	installBin   string
	embedBin     string
	coreDir      string
	repoRoot     string
	referenceApp string
	tmpRoot      string
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		fmt.Fprintln(os.Stderr, "integration setup failed:", err)
		os.Exit(1)
	}
	code := m.Run()
	if tmpRoot != "" {
		os.RemoveAll(tmpRoot)
	}
	os.Exit(code)
}

func setup() error {
	wd, err := os.Getwd() // .../core/cmd/leitwerk
	if err != nil {
		return err
	}
	coreDir = filepath.Clean(filepath.Join(wd, "..", ".."))
	repoRoot = filepath.Dir(coreDir)
	referenceApp = filepath.Join(repoRoot, "examples", "reference-app")

	tmp, err := os.MkdirTemp("", "lw-it-")
	if err != nil {
		return err
	}
	tmpRoot = tmp
	raw := filepath.Join(tmp, "leitwerk-raw")
	build := exec.Command("go", "build", "-o", raw, ".")
	build.Dir = wd
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("go build: %v\n%s", err, out)
	}

	// Full install layout with on-disk assets.
	install := filepath.Join(tmp, "install")
	if err := os.MkdirAll(filepath.Join(install, "bin"), 0o755); err != nil {
		return err
	}
	installBin = filepath.Join(install, "bin", "leitwerk")
	if err := copyFile(raw, installBin, 0o755); err != nil {
		return err
	}
	if err := copyTree(filepath.Join(coreDir, "checks"), filepath.Join(install, "checks")); err != nil {
		return err
	}
	if err := copyTree(filepath.Join(coreDir, "templates"), filepath.Join(install, "templates")); err != nil {
		return err
	}
	if err := copyFile(filepath.Join(coreDir, "leitwerk.tiers"), filepath.Join(install, "leitwerk.tiers"), 0o644); err != nil {
		return err
	}

	// Binary-only layout (no siblings) — forces the embedded fallback.
	embed := filepath.Join(tmp, "embedonly", "bin")
	if err := os.MkdirAll(embed, 0o755); err != nil {
		return err
	}
	embedBin = filepath.Join(embed, "leitwerk")
	return copyFile(raw, embedBin, 0o755)
}

// runBin runs a built binary and returns its exit code and combined streams.
func runBin(t *testing.T, bin, dir string, args ...string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	var out, errb strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	code := 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
		} else {
			t.Fatalf("running %s %v: %v", bin, args, err)
		}
	}
	return code, out.String(), errb.String()
}

func TestIntegrationTierMapping(t *testing.T) {
	// Run from an empty temp dir so no local leitwerk/tiers.conf shadows the
	// shipped defaults next to the binary.
	dir := t.TempDir()
	cases := map[string]string{
		"db/migrations/001.sql": "T2",
		"infra/main.tf":         "T2",
		"docs/guide.md":         "T0",
		"src/app.py":            "T1",
	}
	for path, want := range cases {
		code, out, _ := runBin(t, installBin, dir, "tier", path)
		if code != 0 || strings.TrimSpace(out) != want {
			t.Errorf("tier %s = %q (exit %d), want %q", path, strings.TrimSpace(out), code, want)
		}
	}
}

func TestIntegrationGuard(t *testing.T) {
	dir := t.TempDir()
	if code, _, _ := runBin(t, installBin, dir, "guard", "leitwerk/constitution.md"); code != 3 {
		t.Errorf("guard constitution exit = %d, want 3", code)
	}
	if code, _, _ := runBin(t, installBin, dir, "guard", "/abs/path/to/leitwerk/tiers.conf"); code != 3 {
		t.Errorf("guard abs-suffix exit = %d, want 3", code)
	}
	if code, _, _ := runBin(t, installBin, dir, "guard", "src/app.py"); code != 0 {
		t.Errorf("guard editable exit = %d, want 0", code)
	}
}

func TestIntegrationVerifyReferenceApp(t *testing.T) {
	code, out, _ := runBin(t, installBin, referenceApp, "verify", "--tier", "T0")
	if code != 0 {
		t.Errorf("verify reference-app exit = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "gate: PASS") {
		t.Errorf("expected PASS on reference-app, got:\n%s", out)
	}
}

func TestIntegrationInit(t *testing.T) {
	dir := t.TempDir()
	sentinel := "KEEP ME\n"
	// Pre-seed every scaffolded file so we can prove init's idempotence contract:
	//   overwrite-always: constitution.md, tiers.conf
	//   create-if-absent: CLAUDE.md, the tier rule, the review workflow
	seed := map[string]string{
		"leitwerk/constitution.md":              sentinel,
		"leitwerk/tiers.conf":                   sentinel,
		"CLAUDE.md":                             sentinel,
		".claude/rules/tier-discipline.md":      sentinel,
		".claude/workflows/leitwerk-review.mjs": sentinel,
	}
	for rel, body := range seed {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	code, out, _ := runBin(t, installBin, t.TempDir(), "init", dir)
	if code != 0 {
		t.Fatalf("init exit = %d\n%s", code, out)
	}

	overwritten := []string{"leitwerk/constitution.md", "leitwerk/tiers.conf"}
	preserved := []string{"CLAUDE.md", ".claude/rules/tier-discipline.md", ".claude/workflows/leitwerk-review.mjs"}
	for _, rel := range overwritten {
		if b, _ := os.ReadFile(filepath.Join(dir, rel)); string(b) == sentinel {
			t.Errorf("init did not overwrite %s (must always be written)", rel)
		}
	}
	for _, rel := range preserved {
		if b, _ := os.ReadFile(filepath.Join(dir, rel)); string(b) != sentinel {
			t.Errorf("init clobbered an existing %s (must be create-if-absent)", rel)
		}
	}

	// A fresh init (no pre-seed) scaffolds the review workflow identical to the
	// template it is copied from.
	fresh := t.TempDir()
	if code, out, _ := runBin(t, installBin, t.TempDir(), "init", fresh); code != 0 {
		t.Fatalf("fresh init exit = %d\n%s", code, out)
	}
	got, _ := os.ReadFile(filepath.Join(fresh, ".claude", "workflows", "leitwerk-review.mjs"))
	want, _ := os.ReadFile(filepath.Join(coreDir, "templates", "workflows", "leitwerk-review.mjs"))
	if string(got) != string(want) {
		t.Errorf("scaffolded review workflow differs from template")
	}
}

func TestIntegrationDrift(t *testing.T) {
	// From the repo root (which has leitwerk/specs), drift reports tracked specs
	// and exits 0. This also exercises the wiring that the subcommand runs the
	// built-in drift check.
	code, out, _ := runBin(t, installBin, repoRoot, "drift")
	if code != 0 {
		t.Errorf("drift at repo root exit = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "spec(s) tracked") {
		t.Errorf("drift output = %q, want a 'spec(s) tracked' summary", out)
	}
	// From a dir with no specs, drift skips (exit 2) — it never fakes a pass.
	if code, _, _ := runBin(t, installBin, t.TempDir(), "drift"); code != 2 {
		t.Errorf("drift with no specs exit = %d, want 2 (skip)", code)
	}
}

func TestIntegrationVerifyBadOption(t *testing.T) {
	// An unknown verify option, and --tier with no value, are usage errors (exit 2).
	if code, _, errb := runBin(t, installBin, t.TempDir(), "verify", "--bogus"); code != 2 || !strings.Contains(errb, "unknown verify option") {
		t.Errorf("verify --bogus = exit %d, stderr %q; want exit 2 + 'unknown verify option'", code, errb)
	}
	if code, _, _ := runBin(t, installBin, t.TempDir(), "verify", "--tier"); code != 2 {
		t.Errorf("verify --tier (no value) exit = %d, want 2", code)
	}
}

func TestIntegrationVersion(t *testing.T) {
	code, out, _ := runBin(t, installBin, t.TempDir(), "version")
	if code != 0 || strings.TrimSpace(out) != "leitwerk 0.1.0" {
		t.Errorf("version = %q (exit %d)", strings.TrimSpace(out), code)
	}
}

func TestIntegrationUnknownCommand(t *testing.T) {
	code, _, errb := runBin(t, installBin, t.TempDir(), "frobnicate")
	if code != 2 {
		t.Errorf("unknown command exit = %d, want 2", code)
	}
	if !strings.Contains(errb, "unknown command") {
		t.Errorf("stderr = %q", errb)
	}
}

func TestIntegrationNoArgsPrintsHelp(t *testing.T) {
	// No args = help (like the Bash CLI), never a crash.
	code, out, _ := runBin(t, installBin, t.TempDir())
	if code != 0 {
		t.Errorf("no-args exit = %d, want 0", code)
	}
	if !strings.Contains(out, "Subcommands:") {
		t.Errorf("no-args output does not look like help:\n%s", out)
	}
}

// TestIntegrationEmbedFallback proves a binary with no sibling files can scaffold
// a repo and run the full gate from its embedded assets — the "single static
// binary, independent of repo layout" claim.
func TestIntegrationEmbedFallback(t *testing.T) {
	dir := t.TempDir()
	if code, out, errb := runBin(t, embedBin, dir, "init", "."); code != 0 {
		t.Fatalf("embed init exit = %d\n%s\n%s", code, out, errb)
	}
	if _, err := os.Stat(filepath.Join(dir, "leitwerk", "tiers.conf")); err != nil {
		t.Errorf("embed init did not scaffold tiers.conf: %v", err)
	}
	code, out, _ := runBin(t, embedBin, dir, "verify", "--tier", "T2")
	if code != 0 {
		t.Errorf("embed verify exit = %d, want 0 (all checks skip)\n%s", code, out)
	}
	if !strings.Contains(out, "gate: PASS") {
		t.Errorf("embed verify expected PASS:\n%s", out)
	}
}

// --- small file helpers ---

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		return copyFile(p, target, info.Mode().Perm())
	})
}
