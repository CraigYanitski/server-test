package auth

import (
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
    JWT, err := token.SignedString(tokenSecret)
    if err != nil {
        return "", err
    }
    return JWT, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
    claims := jwt.MapClaims{}
    _, err := jwt.ParseWithClaims(
        tokenString, 
        claims, 
        func(t *jwt.Token) (interface{}, error) {
            return []byte(tokenSecret), nil
        },
    )
    if err != nil {
        return uuid.New(), err
    }
    user, err := claims.GetSubject()
    if err != nil {
        return uuid.New(), err
    }
    return uuid.Parse(user)
}

