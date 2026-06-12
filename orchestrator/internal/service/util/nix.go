package util

import (
	"os/exec"
	"regexp"
)

// NixPkgRe извлекает имя пакета из nix store path: /nix/store/hash-name-version' from
var NixPkgRe = regexp.MustCompile(`/nix/store/[^-]+-(.+?)' from`)

// NixAvailable проверяет доступность nix в системе
func NixAvailable() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}

// NixDevelopCmd оборачивает команду в nix develop, если nix доступен.
// Возвращает готовый к запуску *exec.Cmd с установленным Dir.
func NixDevelopCmd(workDir, shellCmd string) *exec.Cmd {
	if NixAvailable() {
		cmd := exec.Command("nix", "develop", "--command", "sh", "-c", shellCmd)
		cmd.Dir = workDir
		return cmd
	}
	cmd := exec.Command("sh", "-c", shellCmd)
	cmd.Dir = workDir
	return cmd
}
