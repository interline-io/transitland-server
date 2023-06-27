package util

import "encoding/json"

func MakeJsonError(msg string) string {
	a := map[string]string{
		"error": msg,
	}
	jj, _ := json.Marshal(&a)
	return string(jj)
}
