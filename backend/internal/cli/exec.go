package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"backend/internal/model"
)

// ExecLauncher launches AI CLIs as real OS subprocesses, wrapping them in
// `nix develop` when Nix is available so the CLI runs inside the user's
// isolated environment.
type ExecLauncher struct {
	// QuietPeriod is how long stdout must be silent before a reply is
	// considered complete. Defaults to 750ms.
	QuietPeriod time.Duration
}

// Launch implements Launcher.
func (l ExecLauncher) Launch(ctx context.Context, spec LaunchSpec) (Process, error) {
	quiet := l.QuietPeriod
	if quiet == 0 {
		quiet = 750 * time.Millisecond
	}

	args := cliInvocation(spec.CLI)
	var cmd *exec.Cmd
	if nixAvailable() {
		profile := filepath.Join(spec.EnvPath, ".octra", "nix-profile")
		_ = os.MkdirAll(filepath.Dir(profile), 0o755)
		nixArgs := append([]string{"develop",
			"--extra-experimental-features", "nix-command flakes",
			"--profile", profile,
			"--command"}, args...)
		cmd = exec.Command("nix", nixArgs...)
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}
	cmd.Dir = spec.EnvPath

	homeDir := filepath.Join(spec.EnvPath, "home")
	os.MkdirAll(filepath.Join(homeDir, ".config"), 0o755)
	os.MkdirAll(filepath.Join(homeDir, ".local", "share"), 0o755)
	cmd.Env = append(os.Environ(), []string{
		"HOME=" + homeDir,
		"XDG_CONFIG_HOME=" + filepath.Join(homeDir, ".config"),
		"XDG_DATA_HOME=" + filepath.Join(homeDir, ".local", "share"),
	}...)
	cmd.Env = append(cmd.Env, llmEnv(spec.LLM)...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout // fold stderr into the same stream

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start cli: %w", err)
	}

	p := &execProcess{
		cmd:    cmd,
		stdin:  stdin,
		quiet:  quiet,
		output: make(chan []byte, 64),
		done:   make(chan struct{}),
	}
	go p.readLoop(stdout)
	go func() { _ = cmd.Wait(); close(p.done) }()
	return p, nil
}

func nixAvailable() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}

// cliInvocation returns the argv used to start a CLI in a persistent,
// stdin-driven mode.
func cliInvocation(cli model.CLIType) []string {
	switch cli {
	default:
		return []string{string(cli)}
	}
}

// llmEnv translates LLM settings into environment variables understood by the
// common AI CLIs.
func llmEnv(c LLMConfig) []string {
	var env []string
	if c.APIKey != "" {
		env = append(env, "ANTHROPIC_API_KEY="+c.APIKey)
	}
	if c.BaseURL != "" {
		env = append(env, "ANTHROPIC_BASE_URL="+c.BaseURL)
	}
	if c.Model != "" {
		env = append(env, "ANTHROPIC_MODEL="+c.Model)
	}
	return env
}

// execProcess is a running CLI subprocess.
type execProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	quiet  time.Duration
	output chan []byte
	done   chan struct{}

	mu     sync.Mutex
	closed bool
}

func (p *execProcess) readLoop(r io.Reader) {
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			select {
			case p.output <- chunk:
			case <-p.done:
				return
			}
		}
		if err != nil {
			return
		}
	}
}

// Send writes a prompt and collects stdout until it goes quiet.
func (p *execProcess) Send(ctx context.Context, prompt string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return "", fmt.Errorf("process closed")
	}

	// Drain any leftover banner/output before sending.
	p.drain()

	if _, err := io.WriteString(p.stdin, strings.TrimRight(prompt, "\n")+"\n"); err != nil {
		return "", fmt.Errorf("write prompt: %w", err)
	}

	var sb strings.Builder
	timer := time.NewTimer(p.quiet)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return sb.String(), ctx.Err()
		case <-p.done:
			return sb.String(), nil
		case chunk := <-p.output:
			sb.Write(chunk)
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(p.quiet)
		case <-timer.C:
			return strings.TrimSpace(sb.String()), nil
		}
	}
}

// drain discards any buffered output without blocking.
func (p *execProcess) drain() {
	for {
		select {
		case <-p.output:
		default:
			return
		}
	}
}

func (p *execProcess) Alive() bool {
	select {
	case <-p.done:
		return false
	default:
		return true
	}
}

func (p *execProcess) Kill() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	_ = p.stdin.Close()
	if p.cmd.Process != nil {
		return p.cmd.Process.Kill()
	}
	return nil
}

func (p *execProcess) PID() int {
	if p.cmd.Process != nil {
		return p.cmd.Process.Pid
	}
	return 0
}
