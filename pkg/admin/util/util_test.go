package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSHA256Base64(t *testing.T) {
	assert.Len(t, SHA256Base64("string"), 44)
}

func TestHashPassword(t *testing.T) {
	password := "foobar"
	out := HashPassword(password)
	assert.NotEqual(t, out, "")
	assert.True(t, ComparePassword(password, out))
}

func TestComparePassword(t *testing.T) {
	assert.True(t, ComparePassword("test", "$2a$10$t84eU4c/J1gGfsDeL..vv.BoHpK0Go/8EpG4.hoZuhx7ulVNvV1iC"))
	assert.False(t, ComparePassword("foo", "$2a$10$t84eU4c/J1gGfsDeL..vv.BoHpK0Go/8EpG4.hoZuhx7ulVNvV1iC"))
}

func TestGenRandomString(t *testing.T) {
	assert.Len(t, GenRandomString(20), 20)
	assert.NotEqual(t, GenRandomString(10), GenRandomString(10))
}
