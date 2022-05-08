package util

import (
	"math/rand"
	"strings"
)

var alpha = "abcdefghijklmnopqrstuvwxyz"

func RandomString(n uint64) string {
	var s strings.Builder

	for i := uint64(0); i < n; i++ {
		s.WriteByte(alpha[rand.Intn(26)])
	}
	return s.String()
}
