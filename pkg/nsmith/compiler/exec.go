package compiler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/nspcc-dev/neo-go/pkg/nsmith/toolchain"
)

func defaultManager() (*toolchain.Manager, error) {
	return toolchain.NewManager()
}

func compileEnv() []string {
	return augmentCompileEnv(os.Environ())
}

func augmentCompileEnv(base []string) []string {
	env := append([]string(nil), base...)
	if getenv(env, "DOTNET_ROOT") == "" {
		for _, root := range dotnetRoots() {
			if hasDotnetRuntime(root) {
				env = setenv(env, "DOTNET_ROOT", root)
				env = prependPath(env, root)
				break
			}
		}
	}
	if getenv(env, "JAVA_HOME") == "" {
		if home := findJavaHome(); home != "" {
			env = setenv(env, "JAVA_HOME", home)
			env = prependPath(env, filepath.Join(home, "bin"))
		}
	}
	return env
}

func dotnetRoots() []string {
	var roots []string
	for _, root := range []string{"/usr/share/dotnet", "/usr/lib/dotnet"} {
		if _, err := os.Stat(root); err == nil {
			roots = append(roots, root)
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots, filepath.Join(home, ".dotnet"))
	}
	return roots
}

func hasDotnetRuntime(root string) bool {
	entries, err := os.ReadDir(filepath.Join(root, "shared", "Microsoft.NETCore.App"))
	return err == nil && len(entries) > 0
}

func findJavaHome() string {
	if home := os.Getenv("JAVA_HOME"); home != "" {
		return home
	}
	candidates := []string{
		"/usr/lib/jvm/java-21-openjdk",
		"/usr/lib/jvm/java-17-openjdk",
		"/usr/lib/jvm/java-21",
		"/usr/lib/jvm/java-17",
	}
	if java, err := exec.LookPath("java"); err == nil {
		real, err := filepath.EvalSymlinks(java)
		if err == nil {
			home := filepath.Clean(filepath.Join(filepath.Dir(real), ".."))
			candidates = append([]string{home}, candidates...)
		}
	}
	for _, dir := range candidates {
		if _, err := os.Stat(filepath.Join(dir, "bin", "java")); err == nil {
			if major, err := javaMajorVersion(dir); err == nil && major >= 17 {
				return dir
			}
		}
	}
	if entries, err := os.ReadDir("/usr/lib/jvm"); err == nil {
		type candidate struct {
			path  string
			major int
		}
		var found []candidate
		for _, e := range entries {
			dir := filepath.Join("/usr/lib/jvm", e.Name())
			major, err := javaMajorVersion(dir)
			if err != nil || major < 17 {
				continue
			}
			found = append(found, candidate{path: dir, major: major})
		}
		sort.Slice(found, func(i, j int) bool { return found[i].major > found[j].major })
		if len(found) > 0 {
			return found[0].path
		}
	}
	return ""
}

func javaMajorVersion(home string) (int, error) {
	out, err := exec.Command(filepath.Join(home, "bin", "java"), "-version").CombinedOutput()
	if err != nil {
		return 0, err
	}
	s := string(out)
	idx := strings.Index(s, `"`)
	if idx < 0 {
		return 0, fmt.Errorf("unexpected java -version output")
	}
	end := strings.Index(s[idx+1:], `"`)
	if end < 0 {
		return 0, fmt.Errorf("unexpected java -version output")
	}
	ver := s[idx+1 : idx+1+end]
	parts := strings.Split(ver, ".")
	if len(parts) == 0 {
		return 0, fmt.Errorf("unexpected java version %q", ver)
	}
	if parts[0] == "1" && len(parts) > 1 {
		return strconv.Atoi(parts[1])
	}
	return strconv.Atoi(parts[0])
}

func getenv(env []string, key string) string {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return strings.TrimPrefix(e, prefix)
		}
	}
	return ""
}

func setenv(env []string, key, value string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			continue
		}
		out = append(out, e)
	}
	return append(out, prefix+value)
}

func prependPath(env []string, dir string) []string {
	current := getenv(env, "PATH")
	if current == "" {
		return setenv(env, "PATH", dir)
	}
	return setenv(env, "PATH", dir+string(os.PathListSeparator)+current)
}

func runCmd(ctx context.Context, dir string, env []string, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findArtifact(dir, ext string) (string, error) {
	var matches []string
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(p), ext) {
			matches = append(matches, p)
		}
		return nil
	})
	if len(matches) == 0 {
		return "", fmt.Errorf("no %s file found under %s", ext, dir)
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	// Prefer shortest path (usually project root output).
	best := matches[0]
	for _, m := range matches[1:] {
		if len(m) < len(best) {
			best = m
		}
	}
	return best, nil
}

func copyIfDifferent(src, dst string) error {
	if src == dst {
		return nil
	}
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, raw, 0o644)
}

func pythonBin(m *toolchain.Manager, name string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(m.PythonVenvDir(), "Scripts", name+".exe")
	}
	return m.PythonBin(name)
}
