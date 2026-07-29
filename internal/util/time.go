package util

import "time"

// ParseTimestamp tries multiple common timestamp formats and returns the first that parses successfully.
// SQLite via go-sqlite3 may store timestamps in RFC3339 format (with T and Z),
// while CURRENT_TIMESTAMP produces "2006-01-02 15:04:05" format.
func ParseTimestamp(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	formats := []string{
		time.RFC3339,           // "2006-01-02T15:04:05Z07:00"
		"2006-01-02T15:04:05Z", // UTC without timezone offset
		"2006-01-02 15:04:05",  // SQLite CURRENT_TIMESTAMP
		"2006-01-02T15:04:05",  // ISO 8601 without Z
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// ParseTimestampPtr is like ParseTimestamp but returns a *time.Time (nil for empty/invalid strings).
func ParseTimestampPtr(s string) *time.Time {
	t := ParseTimestamp(s)
	if t.IsZero() {
		return nil
	}
	return &t
}