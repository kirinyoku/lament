package random

import (
	"math/rand"
	"time"
)

func NewRandomString(strLength int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "abcdefghijklmnopqrstuvwxyz" + "0123456789")

	buf := make([]rune, strLength)
	for i := range buf {
		buf[i] = chars[rnd.Intn(len(chars))]
	}

	return string(buf)
}
