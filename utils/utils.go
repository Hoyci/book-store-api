package utils

import (
	"time"
)

func StringPtr(s string) *string {
	return &s
}

func IntPtr(i int) *int {
	return &i
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

func IsNil(value any) bool {
	switch v := value.(type) {
	case *string:
		return v == nil
	case *int:
		return v == nil
	case *[]string:
		return v == nil
	default:
		return value == nil
	}
}
