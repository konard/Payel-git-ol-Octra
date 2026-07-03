package cli

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type OcaweLauncher struct{}

type OcaweProcess struct {
	cmd    *exec.Cmd
	port   int
	done   chan struct{}
	mu     sync.Mutex
	closed bool
}

func (l OcaweLauncher) Launch(ctx context.Context, spec LaunchSpec) (Process, error) {
	port, err := findFreePort()
	if err != nil {
		return nil, fmt.Errorf("find free port: %w", err)
	}

	args := []string{
		"--port", fmt.Sprintf("%d", port),
		"--workflows-root", filepath.Join(spec.EnvPath, "workflows"),
	}
	cmd := exec.Command("ocawecore", args...)
	cmd.Dir = spec.EnvPath

	homeDir := filepath.Join(spec.EnvPath, "home")
	os.MkdirAll(filepath.Join(homeDir, ".config"), 0o755)
	os.MkdirAll(filepath.Join(homeDir, ".local", "share"), 0o755)
	cmd.Env = prependPath(os.Environ(), profileBinPaths(spec.EnvPath))
	cmd.Env = append(cmd.Env, []string{
		"HOME=" + homeDir,
		"XDG_CONFIG_HOME=" + filepath.Join(homeDir, ".config"),
		"XDG_DATA_HOME=" + filepath.Join(homeDir, ".local", "share"),
		"OCAWE_PORT=" + fmt.Sprintf("%d", port),
	}...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start ocawecore: %w", err)
	}

	go io.Copy(io.Discard, stderr)

	if err := waitForHealth(ctx, port, 30*time.Second); err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("ocawecore health check: %w", err)
	}

	p := &OcaweProcess{
		cmd:  cmd,
		port: port,
		done: make(chan struct{}),
	}
	go func() { _ = cmd.Wait(); close(p.done) }()
	return p, nil
}

func (p *OcaweProcess) Send(ctx context.Context, prompt string) (string, error) {
	return "", fmt.Errorf("OcaweProcess does not support Send; use HTTP API directly")
}

func (p *OcaweProcess) Alive() bool {
	select {
	case <-p.done:
		return false
	default:
		return true
	}
}

func (p *OcaweProcess) Kill() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	if p.cmd.Process != nil {
		return p.cmd.Process.Kill()
	}
	return nil
}

func (p *OcaweProcess) PID() int {
	if p.cmd.Process != nil {
		return p.cmd.Process.Pid
	}
	return 0
}

func (p *OcaweProcess) Port() int {
	return p.port
}

func findFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port, nil
}

func waitForHealth(ctx context.Context, port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	url := fmt.Sprintf("http://127.0.0.1:%d/health", port)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for ocawecore on port %d", port)
}
