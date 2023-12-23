package types

import (
	"encoding/json"
)

type H map[string]interface{}

func stringUnmarshalJSON(b []byte) (s string, err error) {
	if err = json.Unmarshal(b, &s); err != nil {
		s = ""
	}
	return s, nil
}
