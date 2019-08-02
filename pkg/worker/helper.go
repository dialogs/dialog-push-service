package worker

import (
	"crypto/md5"
	"encoding/hex"
	"io"
)

func MD5(src string) (string, error) {

	h := md5.New()
	_, err := io.WriteString(h, src)
	if err != nil {
		return "", err
	}

	res := hex.EncodeToString(h.Sum(nil))
	return res, nil
}

func TokenHash(src string) string {

	tokenHash, err := MD5(src)
	if err != nil {
		tokenHash = "unknown"
	}

	return tokenHash
}
