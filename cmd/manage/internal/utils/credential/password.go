package credential

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/argon2"
)

// New 为新密码生成随机盐和 Argon2id 摘要。
func New(password string) (digest, salt string, err error) {
	if len(password) < 12 || len(password) > 1024 {
		return "", "", errors.New("管理员密码长度须为 12 至 1024 字节")
	}
	var rawSalt [16]byte
	if _, err := rand.Read(rawSalt[:]); err != nil {
		return "", "", err
	}
	return hash(password, rawSalt[:]), hex.EncodeToString(rawSalt[:]), nil
}

// Matches 校验密码摘要。
func Matches(password, salt, digest string) bool {
	rawSalt, err := hex.DecodeString(salt)
	if err != nil || len(rawSalt) != 16 || len(password) > 1024 {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(hash(password, rawSalt)), []byte(digest)) == 1
}

func hash(password string, salt []byte) string {
	return hex.EncodeToString(argon2.IDKey([]byte(password), salt, 2, 19*1024, 1, 32))
}
