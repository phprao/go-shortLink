package lib

import (
	"crypto/sha1"
	"fmt"
	"io"
)

func ToSha1(str string) string {
	s := sha1.New()
	_, err := io.WriteString(s, str)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", s.Sum(nil))
}