package date

import "time"

func IsExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}

func IsDaysOld(test time.Time) int {
	return int(time.Since(test).Abs().Hours()/24)
}