package worker

import "encoding/json"

func marshalJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
