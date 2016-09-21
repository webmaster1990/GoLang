package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func createApiKey(userId string) string {
	key := []byte(AUTH_SECRET)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(userId))
	hash := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(hash)
}
