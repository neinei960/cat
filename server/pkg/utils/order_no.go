package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateOrderNo generates a unique order number: yyyyMMddHHmmss + 6 random digits
func GenerateOrderNo() string {
	now := time.Now()
	return fmt.Sprintf("%s%06d", now.Format("20060102150405"), rand.Intn(1000000))
}
