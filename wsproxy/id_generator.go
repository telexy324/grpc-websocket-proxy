package wsproxy

import (
	uuid "github.com/iris-contrib/go.uuid"
	"math/rand"
	"time"
)

var (
	DefaultIDGenerator = func() string {
		id, err := uuid.NewV4()
		if err != nil {
			return randomString(64)
		}
		return id.String()
	}
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

// random takes a parameter (int) and returns random slice of byte
// ex: var randomstrbytes []byte; randomstrbytes = utils.Random(32)
func random(n int) []byte {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return b
}

// randomString accepts a number(10 for example) and returns a random string using simple but fairly safe random algorithm
func randomString(n int) string {
	return string(random(n))
}
