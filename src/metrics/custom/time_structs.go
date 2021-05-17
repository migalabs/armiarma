package custom

import (
	"time"
)


// Struct that represents the actual time of a given time
type TimeStruct struct {
	Year int
	Month time.Month
	Day int
	Hour int
	Minute int
	Second int
}

// NewTimeStruct generates a new TimeStruct with the 
// inner fields empty
func NewTimeStruct() TimeStruct {
	ts := TimeStruct {
		Year: 0,
		Month: 0,
		Day: 0,
		Hour: 0,
		Minute: 0,
		Second: 0,
	}
	return ts
}

// StampCurrentTime stamps on the variables of the Struct 
// the current time divided in fields
func (ts *TimeStruct) StampCurrentTime() {
	currentTime := time.Now()
	ts.Year = currentTime.Year()
	ts.Month = currentTime.Month()
	ts.Day = currentTime.Day()
	ts.Hour = currentTime.Hour()
	ts.Minute = currentTime.Minute()
	ts.Second = currentTime.Second()
}