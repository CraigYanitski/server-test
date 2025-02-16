package main

import (
    "net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
    // check if there is admin access
    if !CheckAdmin(w) {
        return
    }
    // reset server hits
    cfg.fileserverHits.Store(0)
    w.WriteHeader(200)
    w.Write([]byte("Hits successfully reset!"))
    // reset users
    cfg.dbQueries.ResetUsers(r.Context())
}

