package utils

import "time"

func FormatDate(date *time.Time) string {
	if date == nil {
		return ""
	}

	return date.Format("2006-01-02 15:04:05")
}
