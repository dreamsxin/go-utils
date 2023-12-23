package types

import (
	"fmt"

	"time"
)

type Jdate string

func (col Jdate) MarshalCSV() (string, error) {
	t, err := time.ParseInLocation("2006-01-02", string(col), cstZone)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("\"%s\"", t.Format("2006-01-02")), nil
}

func (col Jdate) MarshalJSON() ([]byte, error) {
	t, err := time.ParseInLocation("2006-01-02", string(col), cstZone)
	if err != nil {
		return nil, err
	}
	var stamp = fmt.Sprintf("\"%s\"", t.Format("2006-01-02"))
	return []byte(stamp), nil
}

func (col *Jdate) UnmarshalJSON(data []byte) error {
	s, _ := stringUnmarshalJSON(data)
	if s == "" {
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, cstZone) //cstZone
	if err != nil {
		return err
	}
	*col = Jdate(t.Format("2006-01-02"))
	return nil
}
