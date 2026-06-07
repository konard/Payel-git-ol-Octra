package boss

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// gradeTask calls the HTTP grader and returns complexity 1-10 (default 5 on error)
func gradeTask(taskText string) int {
	client := &http.Client{Timeout: 5 * time.Second}
	body := map[string]string{"task": taskText}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "http://octra-grader:50055/grade", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return 5
	}
	defer resp.Body.Close()
	var res struct {
		Grade int `json:"grade"`
	}
	if json.NewDecoder(resp.Body).Decode(&res) == nil && res.Grade >= 1 && res.Grade <= 10 {
		return res.Grade
	}
	// fallback: try to read raw int from body
	raw, _ := io.ReadAll(resp.Body)
	if len(raw) > 0 {
		var g int
		if _, err := fmt.Sscanf(string(raw), "%d", &g); err == nil && g >= 1 && g <= 10 {
			return g
		}
	}
	return 5
}

