package main

import (
    "net/http"
    "os"
)

func  (cfg *apiConfig) CheckAdmin(w http.ResponseWriter) bool {
    if cfg.platform != "dev" {
        respondWithError(w, http.StatusForbidden, "No admin access", nil)
        return false
    }
    return true
}

