package types

import (
	"fmt"

	"time"
)

type Jepoch int64

func (col Jepoch) MarshalCSV() (string, error) {
	return fmt.Sprintf("\"%s\"", time.Time(time.Unix(int64(col), 0)).In(cstZone).Format("2006-01-02 15:04:05")), nil
}

func (col Jepoch) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", time.Time(time.Unix(int64(col), 0)).In(cstZone).Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}

func (col *Jepoch) UnmarshalJSON(data []byte) error {
	s, _ := stringUnmarshalJSON(data)
	if s == "" {
		*col = Jepoch(0)
		return nil
	}
	//t, err := time.Parse("2006-01-02 15:04:05", s)
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, cstZone) //cstZone
	if err != nil {
		return err
	}
	*col = Jepoch(t.Unix())
	return nil
}
