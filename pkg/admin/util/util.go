package util

import (
	"crypto/sha256"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
	"io"
	"math/rand/v2"
)

const (
	randStrChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// SHA256Base64 hashes a string to SHA256 and encodes it to base64.
// It panics on error.
func SHA256Base64(str string) string {
	hash := sha256.New()

	if _, err := io.WriteString(hash, str); err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

// HashPassword returns a bcrypt encoded password hash.
// The password is hashed using SHA256 first to limit its length.
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(SHA256Base64(password)), bcrypt.DefaultCost)

	if err != nil {
		panic(err)
	}

	return string(bytes)
}

// ComparePassword compares the plain text password and the hash and returns true if they're equal or false otherwise.
func ComparePassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(SHA256Base64(password))) == nil
}

// GenRandomString generates a random string with length n for a fixed alphabet.
func GenRandomString(n uint) string {
	randStr := make([]byte, n)

	for i := range randStr {
		randStr[i] = randStrChars[rand.IntN(len(randStrChars))]
	}

	return string(randStr)
}
