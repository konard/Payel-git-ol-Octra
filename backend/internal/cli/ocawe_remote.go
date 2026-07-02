package cli

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

type RemoteOcaweLauncher struct {
	BaseURL *url.URL
}

func (l RemoteOcaweLauncher) Launch(ctx context.Context, spec LaunchSpec) (Process, error) {
	port := 4111
	if l.BaseURL.Port() != "" {
		if p, err := strconv.Atoi(l.BaseURL.Port()); err == nil {
			port = p
		}
	}
	return &RemoteOcaweProcess{
		baseURL: l.BaseURL,
		port:    port,
	}, nil
}

type RemoteOcaweProcess struct {
	baseURL *url.URL
	port    int
}

func (p *RemoteOcaweProcess) Send(ctx context.Context, prompt string) (string, error) {
	return "", fmt.Errorf("RemoteOcaweProcess does not support Send; use HTTP API directly")
}

func (p *RemoteOcaweProcess) Alive() bool {
	return true
}

func (p *RemoteOcaweProcess) Kill() error {
	return nil
}

func (p *RemoteOcaweProcess) PID() int {
	return 0
}

func (p *RemoteOcaweProcess) Port() int {
	return p.port
}
