package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/CraigYanitski/server-test/internal/auth"
	"github.com/CraigYanitski/server-test/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
    ID         uuid.UUID  `json:"id"`
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`
    Body       string     `json:"body"`
    UserID     uuid.UUID  `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
    // type chirpError struct {Error string `json:"error"`}

    // check user authentication
    token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "", err)
        return
    }
    id, err := auth.ValidateJWT(token, cfg.secret)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, token, err)
        return
    }

    // decode request body
    decoder := json.NewDecoder(r.Body)
    chp := &Chirp{}
    err = decoder.Decode(chp)
    if err != nil {
        //fmt.Printf("error decoding a JSON: %s\n", err)
        //w.WriteHeader(500)
        respondWithError(w, http.StatusInternalServerError, "error decoding JSON", err)
        return
    }

    // validate chirp length and sanitize
    var cleanChirp string
    if len(chp.Body) > 140 {
        respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
        return
    } else {
        cleanChirp = CleanChirpBody(chp.Body)
    }

    // create chirp
    params := database.CreateChirpParams{
        Body: cleanChirp, 
        UserID: id,
    }
    chirp, err := cfg.dbQueries.CreateChirp(r.Context(), params)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "error creating chirp", err)
        return
    }
    respondWithJSON(w, http.StatusCreated, Chirp(chirp))
    return
}

func CleanChirpBody(body string) string {
    clean_body := []string{}
    badWords := map[string]struct{} {
        "kerfuffle": {},
        "sharbert":  {},
        "fornax":    {},
    }
    for _, word := range strings.Fields(body) {
        if _, ok := badWords[strings.ToLower(word)]; ok {
                clean_body = append(clean_body, "****")
            } else {
                clean_body = append(clean_body, word)
            }
        }
    return strings.Join(clean_body, " ")
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
    // check if single chirp specified and call appropriate function
    //id := r.PathValue("chirp_id")
    //if id != "" {
    //    chirpID, err := uuid.Parse(id)
    //    if err != nil {
    //        respondWithError(w, http.StatusInternalServerError, "invalid chirp ID", err)
    //        return
    //    }
    //    cfg.getSingleChirp(w, r, chirpID)
    //    return
    //}

    // define slice of chirps otherwise
    items := []Chirp{}

    // check if user is specified
    idQuery := r.URL.Query().Get("author_id")
    if idQuery != "" {
        userID, err := uuid.Parse(idQuery)
        if err != nil {
            respondWithError(w, http.StatusInternalServerError, "unable to convert given user_id to UUID", err)
            return
        }
        userChirps, err := cfg.dbQueries.GetChirpsByUser(r.Context(), userID)
        if err != nil {
            respondWithError(w, http.StatusNotFound, fmt.Sprintf("%s chirps not found", userID), err)
            return
        }
        
        for _, item := range userChirps {
            items = append(items, Chirp(item))
        }
        respondWithJSON(w, http.StatusOK, items)
        return
    }

    // otherwise return chirps ordered by time
    chirps, err := cfg.dbQueries.GetChirps(r.Context())
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "error getting chirp data", err)
        return
    }
    for _, item := range chirps {
        items = append(items, Chirp(item))
    }
    respondWithJSON(w, http.StatusOK, items)
    return
}

func (cfg *apiConfig) getSingleChirp(w http.ResponseWriter, r *http.Request, chirpID uuid.UUID) {
    chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)
    if err != nil {
        respondWithError(w, http.StatusNotFound, "error finding chirp", err)
        return
    }
    respondWithJSON(w, http.StatusOK, Chirp(chirp))
    return
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
    // authenticate user
    token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "missing access token", err)
        return
    }
    userID, err := auth.ValidateJWT(token, cfg.secret)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "invalid JWT", err)
        return
    }

    // get chirp information
    id := r.PathValue("chirp_id")
    var chirpID uuid.UUID
    if id == "" {
        respondWithError(w, http.StatusInternalServerError, "no chirp ID given", nil)
        return
    } else {
        chirpID, err = uuid.Parse(id)
        if err != nil {
            respondWithError(w, http.StatusInternalServerError, "error parsing UUID from chirp ID", err)
            return
        }
    }
    chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)
    if err != nil {
        respondWithError(w, http.StatusNotFound, "chirp not found", err)
        return
    }

    // verify authorisation
    if chirp.UserID != userID {
        respondWithError(w, http.StatusForbidden, "action forbidden: incorrect user_id", nil)
        return
    }

    // remove chirp and respond with success
    _, err = cfg.dbQueries.DeleteChirp(r.Context(), chirpID)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "error deleting chirp", err)
    }
    respondWithJSON(w, http.StatusNoContent, nil)
    return
}

