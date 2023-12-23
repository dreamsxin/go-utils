package types

import (
	"fmt"
	"strconv"
)

type Serial int64

func (col Serial) MarshalCSV() (string, error) {
	return fmt.Sprintf("\"%s\"", strconv.FormatInt(int64(col), 10)), nil
}

func (col Serial) MarshalJSON() ([]byte, error) {

	var stamp = fmt.Sprintf("\"%s\"", strconv.FormatInt(int64(col), 10))
	return []byte(stamp), nil
}

func (col *Serial) UnmarshalJSON(data []byte) error {
	s, _ := stringUnmarshalJSON(data)
	if s == "" {
		*col = Serial(0)
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*col = Serial(v)
	return nil
}
