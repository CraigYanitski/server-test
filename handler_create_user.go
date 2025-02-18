package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/CraigYanitski/server-test/internal/auth"
	"github.com/CraigYanitski/server-test/internal/database"
	"github.com/google/uuid"
)

// user struct to unmarshal POST requests
type InitUser struct {
    Email     string  `json:"email"`
    Password  string  `json:"password"`
}
// user struct to marshal responses
type User struct {
    ID              uuid.UUID  `json:"id"`
    CreatedAt       time.Time  `json:"created_at"`
    UpdatedAt       time.Time  `json:"updated_at"`
    Email           string     `json:"email"`
    HashedPassword  string     `json:"hashed_password,omitempty"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
    // unmarshal the POST JSON and verify required fields are valid
    decoder := json.NewDecoder(r.Body)
    u := &InitUser{}
    err := decoder.Decode(u)
    if (err != nil) || (u.Email == "") || (u.Password == "") {
        respondWithError(
            w, 
            http.StatusInternalServerError, 
            fmt.Sprintf("error decoding JSON with email '%s' and password '%s'", u.Email, u.Password), 
            err,
        )
        return
    }

    // hash given password
    hash, err := auth.HashPassword(u.Password)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "error hashing password", err)
        return
    }

    // add user to database
    params := database.CreateUserParams{
        Email: u.Email, 
        HashedPassword: hash,
    }
    user, err := cfg.dbQueries.CreateUser(r.Context(), params)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "error creating user", err)
    }

    // empty password field to remove from marshalled JSON
    user.HashedPassword = ""
    respondWithJSON(w, http.StatusCreated, User(user))
    return
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
    // unmarshal the POST JSON and verify required fields are valid
    decoder := json.NewDecoder(r.Body)
    u := &InitUser{}
    err := decoder.Decode(u)
    if (err != nil) || (u.Email == "") || (u.Password == "") {
        respondWithError(
            w, 
            http.StatusInternalServerError, 
            fmt.Sprintf("error decoding JSON with email '%s' and password '%s'", u.Email, u.Password), 
            err,
        )
        return
    }

    // search for user in database using their email
    user, err := cfg.dbQueries.GetUserByEmail(r.Context(), u.Email)
    if (err != nil) || (user.HashedPassword == "") {
        respondWithError(w, http.StatusNotFound, "error finding user", err)
    }

    // check validity of password
    err = auth.CheckPasswordHash(u.Password, user.HashedPassword)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "password incorrect", err)
        return
    }

    // empty password field to remove from marshalled JSON
    user.HashedPassword = ""

    // respond with user JSON
    respondWithJSON(w, http.StatusOK, User(user))
}

