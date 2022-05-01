package timex

import (
	"strconv"
	"strings"
	"time"
)

func paddingLeft(s string, n int) string {
	if len(s) >= n {
		return s
	}

	return paddingLeft("0"+s, n)
}

func FormatDuration(d time.Duration) string {
	parts := []string{}

	if d.Hours() > 0 {
		hours := int(d.Hours())
		d = d - time.Duration(hours)*time.Hour
		parts = append(parts, paddingLeft(strconv.Itoa(hours), 2)+"h")
	}

	minutes := int(d.Minutes())
	parts = append(parts, paddingLeft(strconv.Itoa(minutes), 2)+"m")
	d = d - time.Duration(minutes)*time.Minute

	seconds := int(d.Seconds())
	parts = append(parts, paddingLeft(strconv.Itoa(seconds), 2)+"s")
	d = d - time.Duration(seconds)*time.Second

	return strings.Join(parts, ":") + "." + paddingLeft(strconv.Itoa(int(d.Nanoseconds()/1000000)), 3)
}
