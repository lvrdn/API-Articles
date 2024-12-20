package user

import (
	"golang.org/x/crypto/argon2"
	"golang.org/x/exp/rand"
)

func hashPassword(originalPassword, salt string) []byte {
	hashedPass := argon2.IDKey([]byte(originalPassword), []byte(salt), 1, 64*1024, 4, 32)
	result := make([]byte, len(salt))
	copy(result, salt)
	return append(result, hashedPass...)
}

func randStringRunes(n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
