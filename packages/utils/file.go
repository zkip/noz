package utils

import (
	"crypto/md5"
	"fmt"
)

func GenSum(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}
