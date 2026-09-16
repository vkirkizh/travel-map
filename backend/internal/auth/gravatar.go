package auth

import (
	"crypto/md5"
	"encoding/hex"
)

func GravatarURL(email string) string {
	normalized := NormalizeEmail(email)

	hash := md5.Sum([]byte(normalized))

	return "https://www.gravatar.com/avatar/" +
		hex.EncodeToString(hash[:]) +
		"?d=identicon&s=160"
}
