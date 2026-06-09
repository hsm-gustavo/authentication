package date

import "time"

func IsExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}