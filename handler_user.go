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
    Email     string   `json:"email"`
    Password  string   `json:"password"`
    Duration  int64  `json:"expires_in_seconds"`
}
// user struct to recast database user and marshal responses
type User struct {
    ID              uuid.UUID  `json:"id"`
    CreatedAt       time.Time  `json:"created_at"`
    UpdatedAt       time.Time  `json:"updated_at"`
    Email           string     `json:"email"`
    HashedPassword  string     `json:"hashed_password,omitempty"`
}
// valid user with additional access token
type ValidUser struct{
    User
    Token  string  `json:"token"`
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
        return
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

    // determine JWT duration
    var duration time.Duration
    if u.Duration >= int64(time.Hour.Seconds()) || u.Duration == 0 {
        duration = time.Duration(time.Hour.Nanoseconds())
    } else {
        duration = time.Duration(u.Duration*time.Second.Nanoseconds())
    }

    // search for user in database using their email
    foundUser, err := cfg.dbQueries.GetUserByEmail(r.Context(), u.Email)
    if (err != nil) || (foundUser.HashedPassword == "") {
        respondWithError(w, http.StatusNotFound, "error finding user", err)
        return
    }

    // make user JWT token
    token, err := auth.MakeJWT(foundUser.ID, cfg.secret, duration)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "error making JWT token", err)
        return
    }

    // recast database user to validated one, adding JWT
    validUser := &ValidUser{}
    validUser.User = User(foundUser)
    validUser.Token = token

    // check validity of password
    err = auth.CheckPasswordHash(u.Password, validUser.HashedPassword)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "password incorrect", err)
        return
    }

    // empty password field to remove from marshalled JSON
    validUser.HashedPassword = ""

    // respond with user JSON
    respondWithJSON(w, http.StatusOK, validUser)
}

