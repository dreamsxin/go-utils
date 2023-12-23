package types

import (
	"fmt"

	"time"
)

// var cstZone = time.FixedZone("CST", 8*3600)       // 东八
var cstZone, _ = time.LoadLocation("Asia/Shanghai")

func init() {
	time.Local = cstZone
}

type Jtime time.Time

func (col Jtime) MarshalCSV() (string, error) {
	return fmt.Sprintf("\"%s\"", time.Time(col).In(cstZone).Format("2006-01-02 15:04:05")), nil
}

func (col Jtime) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", time.Time(col).In(cstZone).Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}

func (col *Jtime) UnmarshalJSON(data []byte) error {
	s, _ := stringUnmarshalJSON(data)
	if s == "" {
		*col = Jtime(time.Now())
		return nil
	}
	//t, err := time.Parse("2006-01-02 15:04:05", s)
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, cstZone) //cstZone
	if err != nil {
		return err
	}
	*col = Jtime(t)
	return nil
}
