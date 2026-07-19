// Command leitwerk is the tool-agnostic verification gate. It is a drop-in
// replacement for the historical Bash script at core/bin/leitwerk, preserving the
// same subcommands, exit codes, and output contract. CI runs it as the
// authoritative gate; Claude Code runs it via a Stop hook and the Bash tool;
// open-code agents run it from AGENTS.md instructions — all three invoke this same
// binary, so the strongest guarantee never depends on a particular agent.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	assets "github.com/cf-sewe/leitwerk-devkit/core"
	"github.com/cf-sewe/leitwerk-devkit/core/internal/gate"
)

const version = "0.1.0"

const helpText = `leitwerk — the tool-agnostic gate.

This is the one deterministic artifact the framework depends on. It is a plain
executable: no agent runtime is required to run it. CI runs it as the
authoritative gate; Claude Code runs it via a Stop hook and the Bash tool;
open-code agents run it from AGENTS.md instructions. All three invoke the SAME
binary, so the strongest guarantee never depends on a particular agent.

Subcommands:
  verify [--tier T0|T1|T2]   Run the checks selected for a blast-radius tier.
  tier <path>                Print the blast-radius tier for a changed path.
  guard <path>               Exit non-zero if a path is human-owned (see below).
  drift                      Surface spec<->code divergence (does not resolve it).
  init [dir]                 Scaffold leitwerk/ (constitution, tiers) into a repo.
  version

Exit codes: 0 = gate green, 1 = a check failed (gate red), 2 = usage error,
            3 = path is human-owned (guard).
`

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

// run is the testable entrypoint: it returns the process exit code instead of
// calling os.Exit, and writes to the provided streams.
func run(args []string, stdout, stderr io.Writer) int {
	cmd := ""
	var rest []string
	if len(args) > 0 {
		cmd, rest = args[0], args[1:]
	}

	switch cmd {
	case "verify":
		return cmdVerify(rest, loadEnv(), stdout, stderr, colorEnabled(stdout))
	case "tier":
		if len(rest) < 1 {
			return usage(stderr, "usage: leitwerk tier <path>")
		}
		e := loadEnv()
		t, ok := e.tiers.TierForPath(rest[0])
		if !ok {
			t = "T1"
		}
		fmt.Fprintln(stdout, t)
		return 0
	case "guard":
		if len(rest) < 1 {
			return usage(stderr, "usage: leitwerk guard <path>")
		}
		e := loadEnv()
		if glob, ok := e.tiers.HumanOwnedMatch(rest[0]); ok {
			fmt.Fprintf(stderr, "leitwerk: '%s' is human-owned (%s).\n", rest[0], glob)
			fmt.Fprintln(stderr, "          An agent may propose changes but may not edit it. See leitwerk/constitution.md.")
			return 3
		}
		return 0
	case "drift":
		return cmdDrift(loadEnv(), stderr)
	case "init":
		dir := "."
		if len(rest) >= 1 {
			dir = rest[0]
		}
		return cmdInit(dir, resolveSelf(), stdout, stderr)
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "leitwerk %s\n", version)
		return 0
	case "", "help", "--help", "-h":
		fmt.Fprint(stdout, helpText)
		return 0
	default:
		return usage(stderr, fmt.Sprintf("unknown command: %s (try 'leitwerk help')", cmd))
	}
}

// env is the resolved configuration for a run: the parsed tiers file plus a check
// resolver. Mirrors the path/env resolution the Bash CLI did at startup.
type env struct {
	self      string
	tiersPath string
	tiers     *gate.Tiers
	resolver  gate.Resolver
}

func loadEnv() env {
	self := resolveSelf()

	tiersPath := getenv("LEITWERK_TIERS", "leitwerk/tiers.conf")
	if !isFile(tiersPath) {
		tiersPath = filepath.Join(self, "leitwerk.tiers")
	}
	data, err := os.ReadFile(tiersPath)
	if err != nil {
		if embedded, e := assets.FS.ReadFile("leitwerk.tiers"); e == nil {
			data = embedded
			tiersPath = "<embedded>"
		}
	}

	return env{
		self:      self,
		tiersPath: tiersPath,
		tiers:     gate.ParseTiers(data),
		resolver: gate.Resolver{
			LocalDir:   getenv("LEITWERK_CHECKS", "leitwerk/checks"),
			BuiltinDir: filepath.Join(self, "checks"),
			CacheDir:   cacheDir(),
			Assets:     assets.FS,
		},
	}
}

func cmdVerify(args []string, e env, stdout, stderr io.Writer, color bool) int {
	tier := "T1"
	for i := 0; i < len(args); {
		a := args[i]
		switch {
		case a == "--tier":
			if i+1 >= len(args) {
				return usage(stderr, "--tier needs a value")
			}
			tier = args[i+1]
			i += 2
		case strings.HasPrefix(a, "--tier="):
			tier = strings.TrimPrefix(a, "--tier=")
			i++
		default:
			return usage(stderr, "unknown verify option: "+a)
		}
	}
	return gate.RunVerify(gate.VerifyOptions{
		Tier:      tier,
		Tiers:     e.tiers,
		Resolver:  e.resolver,
		TiersPath: e.tiersPath,
		Stdout:    stdout,
		Stderr:    stderr,
		Color:     color,
	})
}

// cmdDrift runs the built-in drift check directly (bypassing the repo-local
// override, matching the historical `exec core/checks/drift.sh`) and propagates
// its exit code. The child inherits the terminal and working directory.
func cmdDrift(e env, stderr io.Writer) int {
	script, ok := e.resolver.BuiltinScript("drift")
	if !ok {
		fmt.Fprintln(stderr, "leitwerk: no drift check available")
		return 1
	}
	c := exec.Command(script)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	err := c.Run()
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	fmt.Fprintf(stderr, "leitwerk: %v\n", err)
	return 1
}

// cmdInit scaffolds a repo. constitution.md and tiers.conf are always written;
// CLAUDE.md, the tier-discipline rule, and the review workflow are written only if
// absent — matching the Bash init's idempotence exactly.
func cmdInit(dir, self string, stdout, stderr io.Writer) int {
	steps := []struct {
		rel       string
		dst       string
		overwrite bool
	}{
		{"templates/constitution.template.md", filepath.Join(dir, "leitwerk", "constitution.md"), true},
		{"leitwerk.tiers", filepath.Join(dir, "leitwerk", "tiers.conf"), true},
		{"templates/CLAUDE.template.md", filepath.Join(dir, "CLAUDE.md"), false},
		{"templates/rules/tier-discipline.md", filepath.Join(dir, ".claude", "rules", "tier-discipline.md"), false},
		{"templates/workflows/leitwerk-review.mjs", filepath.Join(dir, ".claude", "workflows", "leitwerk-review.mjs"), false},
	}
	for _, s := range steps {
		if !s.overwrite && isFile(s.dst) {
			continue
		}
		data, err := readAsset(self, s.rel)
		if err != nil {
			fmt.Fprintf(stderr, "leitwerk: init: %v\n", err)
			return 1
		}
		if err := os.MkdirAll(filepath.Dir(s.dst), 0o755); err != nil {
			fmt.Fprintf(stderr, "leitwerk: init: %v\n", err)
			return 1
		}
		if err := os.WriteFile(s.dst, data, 0o644); err != nil {
			fmt.Fprintf(stderr, "leitwerk: init: %v\n", err)
			return 1
		}
	}
	fmt.Fprintf(stdout, "scaffolded %s/: leitwerk/{constitution.md,tiers.conf}, CLAUDE.md, .claude/rules/tier-discipline.md, .claude/workflows/leitwerk-review.mjs\n", dir)
	return 0
}

// readAsset reads an asset by its slash-separated relative path, preferring an
// on-disk copy next to the binary and falling back to the embedded copy.
func readAsset(self, rel string) ([]byte, error) {
	if data, err := os.ReadFile(filepath.Join(self, filepath.FromSlash(rel))); err == nil {
		return data, nil
	}
	return assets.FS.ReadFile(rel)
}

// cacheDir returns a per-user directory for extracted embedded check scripts.
// Using os.UserCacheDir (under the user's home on every platform) rather than a
// shared, world-writable /tmp path stops another local user from pre-planting an
// executable at a predictable location that the gate would then run.
func cacheDir() string {
	base, err := os.UserCacheDir()
	if err != nil || base == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "leitwerk", version)
}

// resolveSelf returns the install root: the parent of the directory holding the
// executable (<root>/bin/leitwerk -> <root>), with symlinks resolved.
func resolveSelf() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	if resolved, e := filepath.EvalSymlinks(exe); e == nil {
		exe = resolved
	}
	return filepath.Dir(filepath.Dir(exe))
}

func usage(w io.Writer, msg string) int {
	fmt.Fprintf(w, "leitwerk: %s\n", msg)
	return 2
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func isFile(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.Mode().IsRegular()
}

// colorEnabled reports whether ANSI colour should be emitted: only when stdout is
// a terminal and NO_COLOR is unset. Deterministic per environment, and keeps piped
// output clean (the parsed subcommands — tier/version — never colour their output).
func colorEnabled(stdout io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := stdout.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}
