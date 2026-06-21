package seatalk

import (
	"crypto/sha256"
	"encoding/hex"
)

// VerifySignature 按 SeaTalk 的方案校验:signature 应等于 hex(sha256(body + signingSecret))。
func VerifySignature(body []byte, signingSecret, signature string) bool {
	hasher := sha256.New()
	hasher.Write(append(body, []byte(signingSecret)...))
	return signature == hex.EncodeToString(hasher.Sum(nil))
}
