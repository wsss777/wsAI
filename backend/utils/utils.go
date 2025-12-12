package utils

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"

	"strings"
	"time"
)

func GetRandomNumbers(num int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var builder strings.Builder

	builder.Grow(num)

	for i := 0; i < num; i++ {
		digit := r.Intn(10)
		builder.WriteByte(byte('0' + digit))
	}

	return builder.String()
}

func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
