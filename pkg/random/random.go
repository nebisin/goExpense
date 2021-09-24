package random

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Int(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func String(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

func Name() string {
	return String(6)
}

func Email() string {
	return fmt.Sprintf("%s@example.com", String(6))
}

func Password() string {
	return fmt.Sprintf("%s%d", String(6), Int(1000, 9999))
}

func Date() time.Time {
	return time.Date(
		int(Int(2000, 2020)),
		time.Month(int(Int(1, 12))),
		int(Int(1, 28)),
		0,
		0,
		0,
		0,
		time.UTC,
	)
}
