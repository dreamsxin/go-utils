package types

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type LocalTime time.Time

func (t LocalTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	tlt := time.Time(t)
	if tlt.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return tlt, nil
}

func (t *LocalTime) Scan(v interface{}) error {
	if value, ok := v.(time.Time); ok {
		*t = LocalTime(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

func (t *LocalTime) String() string {
	// 如果时间 null 那么我们需要把返回的值进行修改
	if t == nil || t.IsZero() {
		return ""
	}
	return time.Time(*t).Format("2006-01-02 15:04:05")
}

func (t *LocalTime) GetDate() string {
	if t == nil || t.IsZero() {
		return ""
	}
	return time.Time(*t).Format(time.DateOnly)
}

func (t *LocalTime) IsZero() bool {
	return time.Time(*t).IsZero()
}

func (t *LocalTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var err error
	//前端接收的时间字符串
	str := string(data)
	//去除接收的str收尾多余的"
	timeStr := strings.Trim(str, "\"")
	t1, err := time.ParseInLocation("2006-01-02 15:04:05", timeStr, time.Local)
	*t = LocalTime(t1)
	return err
}

func (t LocalTime) MarshalJSON() ([]byte, error) {
	tTime := time.Time(t)
	// 如果时间值是空或者0值 返回为null 如果写空字符串会报错
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", tTime.Format("2006-01-02 15:04:05"))), nil
}

type LocalDate time.Time

func (t LocalDate) Value() (driver.Value, error) {
	var zeroTime time.Time
	tlt := time.Time(t)
	if tlt.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return tlt, nil
}

func (t *LocalDate) Scan(v interface{}) error {
	if value, ok := v.(time.Time); ok {
		*t = LocalDate(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

func (t *LocalDate) String() string {
	// 如果时间 null 那么我们需要把返回的值进行修改
	if t == nil || t.IsZero() {
		return ""
	}
	return time.Time(*t).Format("2006-01-02")
}

func (t *LocalDate) IsZero() bool {
	return time.Time(*t).IsZero()
}

func (t *LocalDate) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var err error
	//前端接收的时间字符串
	str := string(data)
	//去除接收的str收尾多余的"
	timeStr := strings.Trim(str, "\"")
	t1, err := time.ParseInLocation("2006-01-02", timeStr, time.Local)
	*t = LocalDate(t1)
	return err
}

func (t LocalDate) MarshalJSON() ([]byte, error) {
	tTime := time.Time(t)
	// 如果时间值是空或者0值 返回为null 如果写空字符串会报错
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", tTime.Format("2006-01-02"))), nil
}

// 小时
type LocalHour time.Time

func (t LocalHour) Value() (driver.Value, error) {
	var zeroTime time.Time
	tlt := time.Time(t)
	if tlt.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return tlt, nil
}

func (t *LocalHour) Scan(v interface{}) error {
	if value, ok := v.(time.Time); ok {
		*t = LocalHour(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

func (t *LocalHour) String() string {
	// 如果时间 null 那么我们需要把返回的值进行修改
	if t == nil || t.IsZero() {
		return ""
	}
	return time.Time(*t).Format("2006-01-02 15")
}

func (t *LocalHour) IsZero() bool {
	return time.Time(*t).IsZero()
}

func (t *LocalHour) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var err error
	//前端接收的时间字符串
	str := string(data)
	//去除接收的str收尾多余的"
	timeStr := strings.Trim(str, "\"")
	t1, err := time.ParseInLocation("2006-01-02 15", timeStr, time.Local)
	*t = LocalHour(t1)
	return err
}

func (t LocalHour) MarshalJSON() ([]byte, error) {
	tTime := time.Time(t)
	// 如果时间值是空或者0值 返回为null 如果写空字符串会报错
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", tTime.Format("2006-01-02 15"))), nil
}
