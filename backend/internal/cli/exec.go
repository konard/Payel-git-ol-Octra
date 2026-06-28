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

// ExecLauncher launches AI CLIs as real OS subprocesses using the user's
// isolated HOME plus Nix profile bin directories on PATH.
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
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = spec.EnvPath

	homeDir := filepath.Join(spec.EnvPath, "home")
	os.MkdirAll(filepath.Join(homeDir, ".config"), 0o755)
	os.MkdirAll(filepath.Join(homeDir, ".local", "share"), 0o755)
	cmd.Env = prependPath(os.Environ(), profileBinPaths(spec.EnvPath))
	cmd.Env = append(cmd.Env, []string{
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
	provider := strings.ToLower(c.Provider)
	if c.APIKey != "" {
		env = append(env,
			"ANTHROPIC_API_KEY="+c.APIKey,
			"OPENAI_API_KEY="+c.APIKey,
		)
		switch provider {
		case "gemini":
			env = append(env, "GEMINI_API_KEY="+c.APIKey)
		case "deepseek":
			env = append(env, "DEEPSEEK_API_KEY="+c.APIKey)
		case "qwen":
			env = append(env, "DASHSCOPE_API_KEY="+c.APIKey)
		case "kimi":
			env = append(env, "MOONSHOT_API_KEY="+c.APIKey)
		case "grok":
			env = append(env, "XAI_API_KEY="+c.APIKey)
		case "openrouter":
			env = append(env, "OPENROUTER_API_KEY="+c.APIKey)
		}
	}
	if c.BaseURL != "" {
		env = append(env,
			"ANTHROPIC_BASE_URL="+c.BaseURL,
			"OPENAI_BASE_URL="+c.BaseURL,
			"OPENAI_API_BASE="+c.BaseURL,
		)
	}
	if c.Model != "" {
		env = append(env,
			"ANTHROPIC_MODEL="+c.Model,
			"OPENAI_MODEL="+c.Model,
		)
	}
	return env
}

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
