package gate

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Resolver locates the executable script for a check by name. It mirrors the
// historical precedence: a repo-local override wins, then the built-in script on
// disk next to the binary, then the built-in embedded in the binary (extracted to
// a cache dir on first use so a `go install`-ed binary with no sibling checks/
// still runs). This keeps the "consuming repo overrides per check; never edits
// installed core" invariant intact.
type Resolver struct {
	LocalDir   string // repo-local checks dir ($LEITWERK_CHECKS or leitwerk/checks)
	BuiltinDir string // on-disk built-in checks dir (<self>/checks); may be absent
	CacheDir   string // where embedded scripts are extracted on demand
	Assets     fs.FS  // embedded FS rooted so that "checks/<name>.sh" resolves
}

// isExecutableFile reports whether p is a regular file with an exec bit set,
// matching the shell `[ -x ]` test used for check resolution.
func isExecutableFile(p string) bool {
	fi, err := os.Stat(p)
	if err != nil || !fi.Mode().IsRegular() {
		return false
	}
	return fi.Mode().Perm()&0o111 != 0
}

// Resolve returns the path to an executable script for the check `name` and true,
// or ("", false) if the check exists in none of the three locations.
func (r Resolver) Resolve(name string) (string, bool) {
	if r.LocalDir != "" {
		if p := filepath.Join(r.LocalDir, name+".sh"); isExecutableFile(p) {
			return p, true
		}
	}
	if r.BuiltinDir != "" {
		if p := filepath.Join(r.BuiltinDir, name+".sh"); isExecutableFile(p) {
			return p, true
		}
	}
	if p, ok := r.extractBuiltin(name); ok {
		return p, true
	}
	return "", false
}

// extractBuiltin writes the embedded checks/<name>.sh to CacheDir and returns its
// path. Returns false if there is no embedded copy. The script is always
// re-extracted (never trusting a file already at the predictable path) into a
// directory proven to be ours, so the gate cannot be tricked into running an
// attacker-planted script.
func (r Resolver) extractBuiltin(name string) (string, bool) {
	if r.Assets == nil || r.CacheDir == "" {
		return "", false
	}
	data, err := fs.ReadFile(r.Assets, "checks/"+name+".sh")
	if err != nil {
		return "", false
	}
	if err := secureCacheDir(r.CacheDir); err != nil {
		return "", false
	}
	dst := filepath.Join(r.CacheDir, name+".sh")
	if err := writeExecutable(dst, data); err != nil {
		return "", false
	}
	return dst, true
}

// secureCacheDir creates dir (and parents) with 0700 and verifies the final
// component is a real directory, not a symlink pointed somewhere an attacker
// controls. It refuses (returns an error) rather than extract into a suspicious
// path. Combined with a per-user base (os.UserCacheDir), this closes the shared-
// /tmp hijack: the directory is under the user's home and only the user can write
// it, and a symlink swapped in at the leaf is rejected.
func secureCacheDir(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	fi, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir() {
		return fmt.Errorf("cache dir %s is not a regular directory", dir)
	}
	if fi.Mode().Perm() != 0o700 { // tighten if MkdirAll reused a looser pre-existing dir
		if err := os.Chmod(dir, 0o700); err != nil {
			return err
		}
	}
	return nil
}

// writeExecutable writes data to an executable file atomically (temp + rename) so
// a concurrent reader never sees a half-written script.
func writeExecutable(dst string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Chmod(0o755); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, dst)
}

// BuiltinScript returns the path to a built-in check script (on-disk if present,
// else extracted from the embedded copy). Used by the `drift` subcommand, which
// runs the built-in check directly, bypassing the repo-local override.
func (r Resolver) BuiltinScript(name string) (string, bool) {
	if r.BuiltinDir != "" {
		if p := filepath.Join(r.BuiltinDir, name+".sh"); isExecutableFile(p) {
			return p, true
		}
	}
	return r.extractBuiltin(name)
}
