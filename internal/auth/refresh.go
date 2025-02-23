package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func MakeRefreshToken() (string, error) {
    token := make([]byte, 32)
    _, err := rand.Reader.Read(token)
    if err != nil {
        return "", fmt.Errorf("error generating refresh token: %s", err)
    }
    return hex.EncodeToString(token), nil
}

