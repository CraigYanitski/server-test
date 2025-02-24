package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
    a, ok := headers["Authorization"]
    if ok {
        key := strings.TrimSpace(strings.TrimPrefix(a[0], "ApiKey"))
        return key, nil
    }
    return "", fmt.Errorf("error: no API key in header")
}
