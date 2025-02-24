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
    Token         string  `json:"token"`
    RefreshToken  string  `json:"refresh_token"`
}

// token struct
type Token struct {
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
    duration := time.Duration(time.Hour.Nanoseconds())
    fmt.Println(duration)

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

    // generate refresh token
    refreshToken, err := cfg.dbQueries.GetRefreshToken(r.Context(), foundUser.ID)
    if err != nil {
        rtExpiresAt := time.Now().AddDate(0, 0, 60)
        rt, err := auth.MakeRefreshToken()
        if err != nil {
            respondWithError(w, http.StatusInternalServerError, "", err)
            return
        }
        params := database.CreateRefreshTokenParams{Token: rt, UserID: foundUser.ID, ExpiresAt: rtExpiresAt}
        refreshToken, err = cfg.dbQueries.CreateRefreshToken(r.Context(), params)
        if err != nil {
            respondWithError(w, http.StatusInternalServerError, "error creating refresh token", err)
            return
        }
    }

    // recast database user to validated one, adding JWT
    validUser := &ValidUser{}
    validUser.User = User(foundUser)
    validUser.Token = token
    validUser.RefreshToken = refreshToken.Token

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
    return
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
    token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "error missing token", err)
        return
    }
    refreshToken, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), token)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "error invalid entry", err)
        return
    }
    newToken, err := auth.MakeJWT(refreshToken.ID, cfg.secret, time.Hour)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "error unauthorised", err)
        return
    }
    respondWithJSON(w, http.StatusOK, Token{newToken})
    return
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
    token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "", err)
        return
    }
    err = cfg.dbQueries.RevokeRefreshToken(r.Context(), token)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "error token not found", err)
        return
    }
    respondWithJSON(w, http.StatusNoContent, nil)
    return
}

