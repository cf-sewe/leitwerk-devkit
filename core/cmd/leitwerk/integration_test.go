package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	cmd.Env = scrubbedEnv() // hermetic: no ambient LEITWERK_* override leaks in
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

// scrubbedEnv is the parent environment with every LEITWERK_* override removed, so
// integration tests exercise the shipped defaults / local fixtures deterministically
// regardless of ambient config (selftest.sh itself exports LEITWERK_TIERS).
func scrubbedEnv() []string {
	var out []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "LEITWERK_") {
			out = append(out, kv)
		}
	}
	return out
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
	if !strings.Contains(out, "scaffolded") { // the observable init-output contract (main.go)
		t.Errorf("init output missing the 'scaffolded …' banner:\n%s", out)
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
	// The line format is a fixed contract: one line `leitwerk <version>`, exit 0.
	// The version is metadata, not a pinned release literal — assert the shape,
	// not a hard-coded number (which would drift from the release tag).
	code, out, _ := runBin(t, installBin, t.TempDir(), "version")
	if code != 0 {
		t.Fatalf("version exit = %d, want 0", code)
	}
	line := strings.TrimRight(out, "\n")
	if strings.Contains(line, "\n") || !regexp.MustCompile(`^leitwerk \S+$`).MatchString(line) {
		t.Errorf("version = %q, want one line matching `^leitwerk \\S+$`", line)
	}

	// installBin is built by setup() with a plain `go build` (no injection), so
	// it must report the `dev` sentinel — a non-release build never claims a tag.
	if line != "leitwerk dev" {
		t.Errorf("default build version = %q, want %q (no ldflags injection)", line, "leitwerk dev")
	}

	// Injection round-trip: a binary built with -ldflags "-X main.version=vTEST"
	// reports exactly `leitwerk vTEST`. This is the release seam the GoReleaser-free
	// release-please.yml relies on; it works only because `version` is a var, not a const.
	injected := filepath.Join(t.TempDir(), "leitwerk-injected")
	build := exec.Command("go", "build", "-ldflags", "-X main.version=vTEST", "-o", injected, ".")
	build.Dir = filepath.Join(coreDir, "cmd", "leitwerk")
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	if b, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build with ldflags injection: %v\n%s", err, b)
	}
	if code, out, _ := runBin(t, injected, t.TempDir(), "version"); code != 0 || strings.TrimSpace(out) != "leitwerk vTEST" {
		t.Errorf("injected version = %q (exit %d), want %q", strings.TrimSpace(out), code, "leitwerk vTEST")
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

// checksLine returns the space-separated check list from verify's "checks:" line
// (the documented contract of which checks a tier runs), or "" if absent.
func checksLine(out string) string {
	for _, ln := range strings.Split(out, "\n") {
		if strings.HasPrefix(ln, "checks:") {
			return strings.TrimSpace(strings.TrimPrefix(ln, "checks:"))
		}
	}
	return ""
}

// TestIntegrationVerifyChecksLine pins the observable checks-line contract: the
// built binary, against the shipped defaults, lists exactly the tier's checks —
// cumulative and in file order. TestChecksForTier covers the function; this
// covers what the user actually sees, and catches a reorder/trim of the shipped
// [tiers] table that the function test on synthetic data would not.
func TestIntegrationVerifyChecksLine(t *testing.T) {
	dir := t.TempDir() // empty + LEITWERK_* scrubbed: shipped defaults next to installBin
	// Each tier's own line, verbatim and cumulative (T1/T2 are independent lines in
	// the file — cumulativeness is a convention ChecksForTier does not enforce — so
	// each is pinned separately; T1 is the default, most-used tier).
	cases := map[string]string{
		"T0": "lint",
		"T1": "lint types tests drift",
		"T2": "lint types tests drift sast erosion",
	}
	for tier, want := range cases {
		// The checks: line is printed before any check runs, so it is deterministic
		// regardless of which checks skip; the run's exit code is not asserted here
		// (it depends on which analysers happen to be installed).
		_, out, errb := runBin(t, installBin, dir, "verify", "--tier", tier)
		if got := checksLine(out); got != want {
			t.Errorf("verify %s checks line = %q, want %q\n%s%s", tier, got, want, out, errb)
		}
	}
}

// TestIntegrationTierFallbackT1 drives main.go's `if !ok { t = "T1" }` fallback,
// unreachable with the shipped catch-all `*`: a [paths] table with no catch-all
// leaves an unmatched path classified T1 (the documented default).
func TestIntegrationTierFallbackT1(t *testing.T) {
	dir := t.TempDir()
	conf := filepath.Join(dir, "leitwerk", "tiers.conf")
	if err := os.MkdirAll(filepath.Dir(conf), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(conf, []byte("[paths]\ndocs/** = T0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// `x.tf` matches the SHIPPED **/*.tf (T2) but NOT this fixture (docs/** only), so
	// a result of T1 proves BOTH that the local fixture loaded AND that main.go's
	// no-match fallback fired. A regression that ignored the fixture — or an ambient
	// LEITWERK_TIERS — would yield T2, making this a discriminating oracle rather
	// than one that passes under the shipped catch-all too.
	if code, out, errb := runBin(t, installBin, dir, "tier", "x.tf"); code != 0 || strings.TrimSpace(out) != "T1" {
		t.Errorf("no-catch-all fallback tier x.tf = %q (exit %d), want T1\n%s", strings.TrimSpace(out), code, errb)
	}
}

// TestIntegrationVerifyAuto exercises `verify --auto`: tier derivation from a git
// diff (the roadmap's acceptance — docs-only → T0, a migration → T2) plus the
// usage-error paths that must never silently under-select a tier.
func TestIntegrationVerifyAuto(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	gitEnv := append(scrubbedEnv(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	git := func(args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
		c.Dir = dir
		c.Env = gitEnv
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	writeFile(t, dir, "docs/readme.md", "# base\n")
	git("add", "-A")
	git("commit", "-qm", "base")
	rp := exec.Command("git", "rev-parse", "HEAD")
	rp.Dir = dir
	rp.Env = gitEnv
	baseOut, err := rp.Output()
	if err != nil {
		t.Fatalf("rev-parse: %v", err)
	}
	base := strings.TrimSpace(string(baseOut))

	// A docs-only change since base derives T0. Asserting the file count too keeps
	// this a real oracle (an empty diff would also print T0).
	writeFile(t, dir, "docs/more.md", "more\n")
	git("add", "-A")
	git("commit", "-qm", "docs")
	if _, out, errb := runBin(t, installBin, dir, "verify", "--auto", "--base", base); !strings.Contains(out, "auto: tier T0") || !strings.Contains(out, "1 changed file(s)") {
		t.Errorf("docs-only --auto: want 'auto: tier T0' with 1 changed file\n%s%s", out, errb)
	}

	// Adding a migration raises the derived tier to T2 (the migration decides), and
	// --auto must actually RUN that tier — assert the checks line is the T2 list.
	writeFile(t, dir, "db/migrations/001_x.sql", "create table x;\n")
	git("add", "-A")
	git("commit", "-qm", "migration")
	_, out, errb := runBin(t, installBin, dir, "verify", "--auto", "--base", base)
	if !strings.Contains(out, "auto: tier T2") {
		t.Errorf("migration --auto: want 'auto: tier T2'\n%s%s", out, errb)
	}
	if got := checksLine(out); got != "lint types tests drift sast erosion" {
		t.Errorf("migration --auto ran checks %q, want the T2 list\n%s", got, out)
	}

	// Working-tree mode (no --base): an UNTRACKED migration is seen via
	// `git ls-files --others` and derives T2 — the path the local Stop hook runs,
	// which the --base cases above never exercise.
	writeFile(t, dir, "db/migrations/003_untracked.sql", "create table z;\n") // not git-added
	if _, out, errb := runBin(t, installBin, dir, "verify", "--auto"); !strings.Contains(out, "auto: tier T2") {
		t.Errorf("working-tree --auto (untracked migration): want 'auto: tier T2'\n%s%s", out, errb)
	}

	// Usage errors (exit 2), each with a DISCRIMINATING message so a guard swap
	// that still exits 2 for a different reason is caught. A precondition failure
	// never under-selects a tier.
	for _, c := range []struct {
		name, wantErr string
		args          []string
		dir           string
	}{
		{"--auto with --tier", "mutually exclusive", []string{"verify", "--auto", "--tier", "T2"}, dir},
		{"option-like base", "looks like an option", []string{"verify", "--auto", "--base", "--output=INJECTED"}, dir},
		{"non-commit base", "not a resolvable commit", []string{"verify", "--auto", "--base", "deadbeef"}, dir},
		{"outside a git work tree", "not inside a git work tree", []string{"verify", "--auto"}, t.TempDir()},
	} {
		code, _, errb := runBin(t, installBin, c.dir, c.args...)
		if code != 2 {
			t.Errorf("%s: exit = %d, want 2\n%s", c.name, code, errb)
		}
		if !strings.Contains(errb, c.wantErr) {
			t.Errorf("%s: stderr = %q, want it to contain %q", c.name, errb, c.wantErr)
		}
	}
	// The option-like base must be refused BEFORE git runs — no leaked --output file.
	if _, err := os.Stat(filepath.Join(dir, "INJECTED")); err == nil {
		t.Errorf("option-like base leaked to git: an INJECTED file was created")
	}
}

// TestIntegrationVerifyAutoNoCommits covers the onboarding path: a fresh repo with
// no HEAD must still derive a tier from the working tree (never block with exit 2).
func TestIntegrationVerifyAutoNoCommits(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	c := exec.Command("git", "init", "-q")
	c.Dir = dir
	c.Env = scrubbedEnv()
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	writeFile(t, dir, "db/migrations/001.sql", "create table x;\n") // untracked, no commits
	if code, out, errb := runBin(t, installBin, dir, "verify", "--auto"); code == 2 || !strings.Contains(out, "auto: tier T2") {
		t.Errorf("no-commits --auto: exit %d, want a T2 derivation (not a block)\n%s%s", code, out, errb)
	}
}

// --- small file helpers ---

func writeFile(t *testing.T, dir, rel, body string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

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
