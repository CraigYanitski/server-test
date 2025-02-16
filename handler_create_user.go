package main

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/google/uuid"
)

type User struct {
    ID         uuid.UUID  `json:"id"`
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`
    Email      string     `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
    decoder := json.NewDecoder(r.Body)
    u := &User{}
    err := decoder.Decode(u)
    if (err != nil) || (u.Email == "") {
        respondWithError(w, http.StatusInternalServerError, "error decoding JSON", err)
        return
    }
    user, err := cfg.dbQueries.CreateUser(r.Context(), u.Email)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "error creating user", err)
    }
    respondWithJSON(w, http.StatusCreated, User(user))
    return
}

