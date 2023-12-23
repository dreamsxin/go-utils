package types

import (
	"fmt"
	"strconv"
)

type Sint32 int32

// Sint32
func (col Sint32) MarshalCSV() (string, error) {
	return fmt.Sprintf("\"%s\"", strconv.FormatInt(int64(col), 10)), nil
}

func (col Sint32) MarshalJSON() ([]byte, error) {

	var stamp = fmt.Sprintf("\"%s\"", strconv.FormatInt(int64(col), 10))
	return []byte(stamp), nil
}

func (col *Sint32) UnmarshalJSON(data []byte) error {
	s, _ := stringUnmarshalJSON(data)
	if s == "" {
		*col = Sint32(0)
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return err
	}
	*col = Sint32(v)
	return nil
}
