package main

import (
	"encoding/json"
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
    id := r.PathValue("chirp_id")
    if id != "" {
        chirp_id, err := uuid.Parse(id)
        if err != nil {
            respondWithError(w, http.StatusInternalServerError, "invalid chirp ID", err)
            return
        }
        cfg.getSingleChirp(w, r, chirp_id)
        return
    }
    chirps, err := cfg.dbQueries.GetChirps(r.Context())
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "error getting chirp data", err)
        return
    }
    items := []Chirp{}
    for _, item := range chirps {
        items = append(items, Chirp(item))
    }
    respondWithJSON(w, http.StatusOK, items)
    return
}

func (cfg *apiConfig) getSingleChirp(w http.ResponseWriter, r *http.Request, chirp_id uuid.UUID) {
    chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirp_id)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "error finding chirp", err)
        return
    }
    respondWithJSON(w, http.StatusOK, Chirp(chirp))
    return
}

