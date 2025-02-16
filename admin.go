package main

import (
    "net/http"
    "os"
)

func CheckAdmin(w http.ResponseWriter) bool {
    platform := os.Getenv("PLATFORM")
    if platform != "dev" {
        respondWithError(w, http.StatusForbidden, "No admin access", nil)
        return false
    }
    return true
}

