package utils

import (
	"encoding/json"
	"fmt"
)

func PrettyJson(object interface{}) string {
	b, err := json.MarshalIndent(object, "", "    ")
	if err != nil {
		fmt.Printf("ERROR: PrettyJson, %v\n %s\n", err, b)
	}
	return string(b)
}
