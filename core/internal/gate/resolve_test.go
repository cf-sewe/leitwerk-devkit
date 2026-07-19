package gate

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func writeScript(t *testing.T, path string, perm os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), perm); err != nil {
		t.Fatal(err)
	}
}

func TestResolvePrecedence(t *testing.T) {
	local := t.TempDir()
	builtin := t.TempDir()
	cache := filepath.Join(t.TempDir(), "cache")
	embed := fstest.MapFS{
		"checks/embedonly.sh": &fstest.MapFile{Data: []byte("#!/bin/sh\nexit 0\n")},
		"checks/lint.sh":      &fstest.MapFile{Data: []byte("#!/bin/sh\nexit 0\n")},
	}
	r := Resolver{LocalDir: local, BuiltinDir: builtin, CacheDir: cache, Assets: embed}

	// Local override wins over built-in and embed.
	writeScript(t, filepath.Join(local, "lint.sh"), 0o755)
	writeScript(t, filepath.Join(builtin, "lint.sh"), 0o755)
	if got, ok := r.Resolve("lint"); !ok || got != filepath.Join(local, "lint.sh") {
		t.Errorf("Resolve(lint) = (%q, %v), want local override", got, ok)
	}

	// Built-in on disk wins when there is no local override.
	writeScript(t, filepath.Join(builtin, "tests.sh"), 0o755)
	if got, ok := r.Resolve("tests"); !ok || got != filepath.Join(builtin, "tests.sh") {
		t.Errorf("Resolve(tests) = (%q, %v), want built-in on disk", got, ok)
	}

	// Falls back to the embedded copy, extracted into the cache dir and executable.
	got, ok := r.Resolve("embedonly")
	if !ok || got != filepath.Join(cache, "embedonly.sh") {
		t.Fatalf("Resolve(embedonly) = (%q, %v), want extracted embed", got, ok)
	}
	if !isExecutableFile(got) {
		t.Errorf("extracted embedded check is not executable: %s", got)
	}

	// Missing everywhere -> not found.
	if _, ok := r.Resolve("nope"); ok {
		t.Errorf("Resolve(nope) should not resolve")
	}
}

func TestResolveSkipsNonExecutableLocal(t *testing.T) {
	local := t.TempDir()
	builtin := t.TempDir()
	r := Resolver{LocalDir: local, BuiltinDir: builtin}
	// A repo-local script without the exec bit is ignored (matches shell `[ -x ]`),
	// so resolution falls through to the built-in.
	writeScript(t, filepath.Join(local, "shell.sh"), 0o644)
	writeScript(t, filepath.Join(builtin, "shell.sh"), 0o755)
	if got, ok := r.Resolve("shell"); !ok || got != filepath.Join(builtin, "shell.sh") {
		t.Errorf("Resolve(shell) = (%q, %v), want built-in (local not executable)", got, ok)
	}
}

func TestBuiltinScriptBypassesLocal(t *testing.T) {
	local := t.TempDir()
	builtin := t.TempDir()
	r := Resolver{LocalDir: local, BuiltinDir: builtin}
	// The drift subcommand uses BuiltinScript, which must ignore a local override.
	writeScript(t, filepath.Join(local, "drift.sh"), 0o755)
	writeScript(t, filepath.Join(builtin, "drift.sh"), 0o755)
	if got, ok := r.BuiltinScript("drift"); !ok || got != filepath.Join(builtin, "drift.sh") {
		t.Errorf("BuiltinScript(drift) = (%q, %v), want built-in only", got, ok)
	}
}
