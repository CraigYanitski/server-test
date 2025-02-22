package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
    token := jwt.NewWithClaims(
        jwt.SigningMethodHS256, 
        jwt.RegisteredClaims{
            Issuer: "chirpy",
            IssuedAt: jwt.NewNumericDate(time.Now()),
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
            Subject: userID.String(),
        },
    )
    JWT, err := token.SignedString([]byte(tokenSecret))
    if err != nil {
        return "", err
    }
    return JWT, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
    claims := &jwt.MapClaims{}
    token, err := jwt.ParseWithClaims(
        tokenString, 
        claims, 
        func(token *jwt.Token) (interface{}, error) {
            return []byte(tokenSecret), nil
        },
    )
    if err != nil {
        return uuid.New(), fmt.Errorf("error parsing JWT during validation: %s", err)
    }
    if !token.Valid {
        return uuid.New(), fmt.Errorf("error: invalid token")
    }
    user, err := claims.GetSubject()
    if err != nil {
        return uuid.New(), fmt.Errorf("error getting claims during validation: %s", err)
    }
    return uuid.Parse(user)
}

func GetBearerToken(headers http.Header) (string, error) {
    a, ok := headers["Authorization"]
    if ok {
        token := strings.TrimSpace(strings.Trim(a[0], "Bearer"))
        return token, nil
    }
    return "", fmt.Errorf("error: no JWT in header")
}

