package worker

import "strings"

// stripMarkdownCodeBlock убирает markdown обёртку ```go ... ``` или ``` ... ```
func stripMarkdownCodeBlock(s string) string {
	s = strings.TrimSpace(s)
	for _, marker := range []string{"```go\n", "```ts\n", "```js\n", "```tsx\n", "```jsx\n", "```json\n", "```yaml\n", "```yml\n", "```sh\n", "```bash\n", "```python\n", "```py\n", "```\n", "```"} {
		if strings.HasPrefix(s, marker) {
			s = s[len(marker):]
			break
		}
	}
	// Remove closing ```
	if end := strings.LastIndex(s, "```"); end > 0 {
		s = strings.TrimSpace(s[:end])
	}
	return s
}
