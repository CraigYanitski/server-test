package auth

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTCreation(t *testing.T) {
    id := uuid.New()
    secret := "Secret token"
    duration := time.Second
    jwt, err := MakeJWT(id, secret, duration)
    if err != nil {
        t.Fatalf("error making JWT: %s", err)
    }
    if strings.Contains(jwt, secret) {
        t.Fatalf("JWT exposes token secret: %s", err)
    }
}

func TestJWTValidation(t *testing.T) {
    id := uuid.New()
    secret := "New secret token"
    duration := time.Second
    jwt, err := MakeJWT(id, secret, duration)
    if err != nil {
        t.Fatalf("error making JWT: %s", err)
    }
    id_validated, err := ValidateJWT(jwt, secret)
    if err != nil {
        t.Fatalf("error in JWT validation: %s", err)
    }
    if id_validated != id {
        t.Fatalf("JWT uuid not same as original: %s", err)
    }
}

func TestJWTValidationFail(t *testing.T) {
    id := uuid.New()
    secret := "Another secret token"
    duration := time.Second
    jwt, err := MakeJWT(id, secret, duration)
    if err != nil {
        t.Fatalf("error making JWT: %s", err)
    }
    wrongSecret := "Another incorrect secret token"
    _, err = ValidateJWT(jwt, wrongSecret)
    if err == nil {
        t.Fatalf("error in JWT validation: it passed when it should have failed...")
    }
}

func TestBearerToken(t *testing.T) {
    header := http.Header{}
    header["Authorization"] = []string{"JWT token here"}
    _, err := GetBearerToken(header)
    if err != nil {
        t.Fatal("error: did not find access token in HTTP header")
    }
}

func TestBearerTokenFail(t *testing.T) {
    header := http.Header{}
    _, err := GetBearerToken(header)
    if err == nil {
        t.Fatal("error: found access token when not supplied")
    }
}

