package request

import (
	"fmt"
	"time"
)

const (
	timeFormat = "2006-01-02 15:04:05"
)

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("Mon Jan _2"))
	return []byte(stamp), nil
}

func (t *JSONTime) UnmarshalJSON(data []byte) (err error) {
	newTime, err := time.ParseInLocation("\""+timeFormat+"\"", string(data), time.Local)
	*t = JSONTime(newTime)
	return
}

// string method
func (t JSONTime) String() string {
	return time.Time(t).Format(timeFormat)
}
