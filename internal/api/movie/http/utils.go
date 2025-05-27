package http

import (
	"time"
)

func parseTime(s string) (parsedTime time.Time, err error) {
	layout := "2006-01-02"
	parsedTime, err = time.Parse(layout, s)

	if err != nil {
		return
	}
	return
}
