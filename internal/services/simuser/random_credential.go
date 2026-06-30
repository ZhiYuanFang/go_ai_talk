package simuser

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	simAccountCharset  = "abcdefghijklmnopqrstuvwxyz0123456789_"
	simPasswordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	simAccountBodyLen  = 11 // 前缀 s + 11 = 12 位，满足 4–32
	simPasswordLen     = 14
)

// GenerateRandomSimCredentials 生成满足 device 用户名/密码规则的随机凭据。
func GenerateRandomSimCredentials() (account, password string, err error) {
	account, err = randomSimAccount()
	if err != nil {
		return "", "", err
	}
	password, err = randomSimPassword()
	if err != nil {
		return "", "", err
	}
	return account, password, nil
}

func randomSimAccount() (string, error) {
	body, err := randomFromCharset(simAccountCharset, simAccountBodyLen)
	if err != nil {
		return "", err
	}
	return "s" + body, nil
}

func randomSimPassword() (string, error) {
	return randomFromCharset(simPasswordCharset, simPasswordLen)
}

func randomFromCharset(charset string, n int) (string, error) {
	if n <= 0 || len(charset) == 0 {
		return "", nil
	}
	max := big.NewInt(int64(len(charset)))
	var b strings.Builder
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b.WriteByte(charset[idx.Int64()])
	}
	return b.String(), nil
}
