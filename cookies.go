package sessions

import (
	"encoding/binary"
	"encoding/base64"
	"crypto/rand"
	"errors"
)

func newCookieAuth(b []byte) string {
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return string(b)
}

func encodeCookie(num uint32, b []byte) string {
	binary.BigEndian.PutUint32(b,num)
	return base64.StdEncoding.EncodeToString(b)
}

func decodeCookie(value string) (uint32, string, error) {
	b, err := base64.StdEncoding.DecodeString(value)
	if err != nil || len(b) < 4 {
		return 0, "", errors.New("Bad cookie")
	}
	num := binary.BigEndian.Uint32(b[:4])
	auth := string(b[4:])
	return num, auth, nil
}