package reviewboard

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"strings"
)

// VerifySignature 校验 Review Board 的 X-Hub-Signature。
// RB 用 HMAC-SHA1(secret, body),header 形如 "sha1=<hexdigest>"——和 SeaTalk 的 sha256(body+secret) 不是一回事。
func VerifySignature(body []byte, secret, header string) bool {
	const prefix = "sha1="
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	want, err := hex.DecodeString(strings.TrimPrefix(header, prefix))
	if err != nil {
		return false
	}
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(body)
	return hmac.Equal(want, mac.Sum(nil)) // 定长比较,避免时序侧信道
}
