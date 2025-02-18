package auth

import (
    "testing"
)

func TestHash(t *testing.T) {
    pw := "some password"
    hash, err := HashPassword(pw)
    if err != nil {
        t.Fatalf("error encountered in hashing password: %s", err)
    }
    if pw == hash {
        t.Fatalf("password '%s' improperly hashed to '%s', cannot be equal", pw, hash)
    }
}

func TestCompare(t *testing.T) {
    pw := "some other password"
    hash, err := HashPassword(pw)
    if err != nil {
        t.Fatalf("error encountered in hashing password: %s", err)
    }
    if err = CheckPasswordHash(pw, hash); err != nil {
        t.Fatalf("password '%s' hash not verified to be '%s', need reproducibility", pw, hash)
    }
}

