package job_schedule

import (
	"math/rand"
	"time"
)

func GetRandomRunAfter(interval float64) time.Time {
	offset := int(interval) + rand.Intn(int(interval))
	return time.Now().Add(time.Duration(rand.Intn(offset)) * time.Second)
}
