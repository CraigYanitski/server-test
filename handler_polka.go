package main

import (
	"encoding/json"
	"net/http"

	"github.com/CraigYanitski/server-test/internal/auth"
	"github.com/google/uuid"
)

type PolkaJSON struct {
    Event  string  `json:"event"`
    Data   any     `json:"data"`
}

type PolkaData struct {
    UserID  uuid.UUID  `json:"user_id"`
}

func (cfg *apiConfig) handlerUpgradeUserToRed(w http.ResponseWriter, r *http.Request) {
    // verify webhook
    key, err := auth.GetAPIKey(r.Header)
    // fmt.Println(key, cfg.polkaKey)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "missing API key", err)
        return
    } else if key != cfg.polkaKey {
        respondWithError(w, http.StatusUnauthorized, "unauthorized API key", err)
        return
    }

    // unmarshal JSON data
    data := &PolkaData{}
    polkaPost := &PolkaJSON{Data: data}
    decoder := json.NewDecoder(r.Body)
    err = decoder.Decode(polkaPost)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "unable to unmarshal JSON", err)
        return
    }

    // verify user upgrade event
    if polkaPost.Event != "user.upgraded" {
        respondWithJSON(w, http.StatusNoContent, nil)
        return
    }

    // upgrade user in DB
    _, err = cfg.dbQueries.UpdateUserToRed(r.Context(), data.UserID)
    if err != nil {
        respondWithError(w, http.StatusNotFound, "unable to find user", err)
        return
    }

    // do not respond with data
    respondWithJSON(w, http.StatusNoContent, nil)
    return
}

