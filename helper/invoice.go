package helper

import (
	"fmt"
	"math/rand"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func GenerateOrderNumberRedis() string {
	return ""
}

func GeneraeteOrderNumber() string {
	prefix := time.Now().Format("20060102")    // YYYYMMDD
	suffix := rng.Intn(900000) + 100000        // 6 digit
	return fmt.Sprintf("%s%d", prefix, suffix) // 16 digit total
}
