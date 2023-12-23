package types

import (
	"strconv"
	"strings"
)

type Int64 int64

func (col *Int64) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(*col), 10)), nil //strconv.Itoa
}

func (col *Int64) UnmarshalJSON(src []byte) error {
	data, _ := strconv.Atoi(strings.Trim(string(src), "\""))
	*col = Int64(data)
	return nil
}
