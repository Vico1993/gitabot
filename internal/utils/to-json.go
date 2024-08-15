package utils

import "encoding/json"

// Output interface in a pretty json in terminal
func ToJson(i interface{}) string {
	bytes, _ := json.MarshalIndent(i, "", "    ")
	return string(bytes)
}
