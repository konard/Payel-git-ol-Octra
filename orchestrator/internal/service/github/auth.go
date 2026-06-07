package github

import (
	"net/url"
	"strings"
)

func (c *Client) authenticatedGitURL(rawURL string) string {
	if c.token == "" || !strings.HasPrefix(rawURL, "https://") {
		return rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	parsed.User = url.UserPassword("x-access-token", c.token)
	return parsed.String()
}

func (c *Client) sanitize(output string) string {
	if c.token == "" {
		return output
	}
	return strings.ReplaceAll(output, c.token, "***")
}

