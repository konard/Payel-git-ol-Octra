package cli

import (
	"os"
	"path/filepath"
	"strings"
)

func profileBinPaths(envPath string) []string {
	baseDir := filepath.Dir(envPath)
	return []string{
		filepath.Join(envPath, ".octra", "nix-profile", "bin"),
		filepath.Join(envPath, "home", ".nix-profile", "bin"),
		filepath.Join(baseDir, ".system", "nix-profile", "bin"),
		filepath.Join(baseDir, ".system", "home", ".nix-profile", "bin"),
	}
}

func prependPath(env []string, dirs []string) []string {
	currentPath := os.Getenv("PATH")
	for _, entry := range env {
		if strings.HasPrefix(entry, "PATH=") {
			currentPath = strings.TrimPrefix(entry, "PATH=")
			break
		}
	}
	pathValue := strings.Join(append(dirs, currentPath), string(os.PathListSeparator))
	next := make([]string, 0, len(env)+1)
	added := false
	for _, entry := range env {
		if strings.HasPrefix(entry, "PATH=") {
			next = append(next, "PATH="+pathValue)
			added = true
			continue
		}
		next = append(next, entry)
	}
	if !added {
		next = append(next, "PATH="+pathValue)
	}
	return next
}
