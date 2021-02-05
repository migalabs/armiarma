package database

import (
    "time"

)

func GetCurrentTime() time.Time {
    return time.Now()
}

func GetTimeMilliseconds() int64 {
    now := time.Now()
    nano := now.UnixNano()
    millis := nano / 1000000 // From nanos to millis 
    return millis
}




