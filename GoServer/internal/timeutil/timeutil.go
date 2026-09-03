package timeutil

import "time"

var IndiaLocation = mustLoadIndiaLocation()

func mustLoadIndiaLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		panic(err)
	}
	return location
}

// Now returns the current time in Indian Standard Time (Asia/Kolkata).
func Now() time.Time {
	return time.Now().In(IndiaLocation)
}
