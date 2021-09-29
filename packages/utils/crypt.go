package utils

import (
	"crypto/md5"
	"encoding/hex"
)

func EncodingStringMd5(plain string) string {
	ctx := md5.New()
	ctx.Write([]byte(plain))

	cipherText := ctx.Sum(nil)

	return hex.EncodeToString(cipherText)
}
