package types

type Json string

func (col Json) MarshalJSON() ([]byte, error) {
	s := string(col)
	if s == "" {
		s = "{}"
	}
	return []byte(s), nil
}

func (col *Json) UnmarshalJSON(data []byte) error {
	*col = Json(data)
	return nil
}
