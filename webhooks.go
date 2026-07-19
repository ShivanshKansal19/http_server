package main

import (
	"encoding/json"
	"net/http"

	"github.com/ShivanshKansal19/http_server/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}
	apiKey, err := auth.GetApiKey(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid API key", err)
		return
	}
	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Invalid API key", nil)
		return
	}
	var params parameters
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	if params.Event == "user.upgraded" {
		_, err = cfg.dbQueries.UpgradeUserToChirpyRed(r.Context(), params.Data.UserID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Couldn't find user", err)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
