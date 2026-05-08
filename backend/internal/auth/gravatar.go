package auth

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

func GravatarURL(email string) string {
	normalized := strings.ToLower(strings.TrimSpace(email))

	hash := md5.Sum([]byte(normalized))

	return "https://www.gravatar.com/avatar/" +
		hex.EncodeToString(hash[:]) +
		"?d=identicon&s=160"
}
