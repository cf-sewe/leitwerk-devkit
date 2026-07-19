package gate

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// colors holds ANSI codes, or empty strings when colouring is off. The off code
// is shared, so a colour-off run collapses every wrapper to plain text.
type colors struct {
	red, grn, yel, dim, off string
}

func newColors(on bool) colors {
	if !on {
		return colors{}
	}
	return colors{
		red: "\x1b[31m",
		grn: "\x1b[32m",
		yel: "\x1b[33m",
		dim: "\x1b[2m",
		off: "\x1b[0m",
	}
}

// VerifyOptions configures a single `verify` run.
type VerifyOptions struct {
	Tier      string
	Tiers     *Tiers
	Resolver  Resolver
	TiersPath string // shown in the "no checks defined" usage error
	Stdout    io.Writer
	Stderr    io.Writer
	Color     bool
}

// RunVerify runs the cumulative check list for a tier and returns the process
// exit code: 0 = green, 1 = a check failed, 2 = usage error (no checks for tier).
// A check's own exit 2 is a clean skip; any other non-zero fails the gate.
func RunVerify(o VerifyOptions) int {
	list := o.Tiers.ChecksForTier(o.Tier)
	if len(list) == 0 {
		fmt.Fprintf(o.Stderr, "leitwerk: no checks defined for tier '%s' in %s\n", o.Tier, o.TiersPath)
		return 2
	}
	c := newColors(o.Color)
	fmt.Fprintf(o.Stdout, "leitwerk verify %s(tier %s)%s\n", c.dim, o.Tier, c.off)
	fmt.Fprintf(o.Stdout, "%schecks: %s%s\n", c.dim, strings.Join(list, " "), c.off)
	fmt.Fprintln(o.Stdout)

	failed := false
	for _, name := range list {
		script, ok := o.Resolver.Resolve(name)
		if !ok {
			fmt.Fprintf(o.Stdout, "  %s?%s  %-10s no such check (looked in %s, %s)\n",
				c.yel, c.off, name, o.Resolver.LocalDir, o.Resolver.BuiltinDir)
			failed = true
			continue
		}
		out, status := runCheck(script)
		last := lastLine(out)
		switch status {
		case 0:
			fmt.Fprintf(o.Stdout, "  %s✓%s  %-10s %s\n", c.grn, c.off, name, last)
		case 2:
			fmt.Fprintf(o.Stdout, "  %s–%s  %-10s %s %s(skipped)%s\n", c.dim, c.off, name, last, c.dim, c.off)
		default:
			fmt.Fprintf(o.Stdout, "  %s✗%s  %-10s %s\n", c.red, c.off, name, last)
			failed = true
		}
	}

	fmt.Fprintln(o.Stdout)
	if !failed {
		fmt.Fprintf(o.Stdout, "%sgate: PASS%s\n", c.grn, c.off)
		return 0
	}
	fmt.Fprintf(o.Stdout, "%sgate: FAIL%s — the change may not land until this is green.\n", c.red, c.off)
	return 1
}

// runCheck executes a check script with the parent working directory and
// environment (checks scan the repo from its root), returning combined
// stdout+stderr and the exit code. A script that cannot start is a failure (126).
func runCheck(script string) (string, int) {
	cmd := exec.Command(script)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return string(out), ee.ExitCode()
	}
	return err.Error(), 126
}

// lastLine returns the text after the final newline, ignoring trailing newlines —
// the equivalent of the shell parameter expansion ${out##*$'\n'} on $(...) output.
func lastLine(s string) string {
	s = strings.TrimRight(s, "\n")
	if i := strings.LastIndexByte(s, '\n'); i >= 0 {
		return s[i+1:]
	}
	return s
}
