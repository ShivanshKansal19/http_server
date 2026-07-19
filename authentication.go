package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ShivanshKansal19/http_server/internal/auth"
	"github.com/ShivanshKansal19/http_server/internal/database"
)

func (cfg *apiConfig) handlerLoginUser(w http.ResponseWriter, r *http.Request) {
	const defaultTokenExpiry = 3600 // 1 hour in seconds
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	var params parameters
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get user by email", err)
		return
	}
	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't check password hash", err)
		return
	}
	if !match {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials", nil)
		return
	}
	token, err := auth.MakeJWT(user.ID, cfg.secretKey, 1*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT", err)
		return
	}
	refreshToken := auth.MakeRefreshToken()
	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token", err)
		return
	}
	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (apiCfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}
	bearerRefreshToken, err := auth.GetBearerToken(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid token", err)
		return
	}
	refreshToken, err := apiCfg.dbQueries.GetRefreshToken(r.Context(), bearerRefreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}
	if refreshToken.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired", nil)
		return
	}
	revoked := refreshToken.RevokedAt.Valid
	if revoked {
		respondWithError(w, http.StatusUnauthorized, "Refresh token revoked", nil)
		return
	}
	token, err := auth.MakeJWT(refreshToken.UserID, apiCfg.secretKey, 1*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT", err)
		return
	}
	respondWithJSON(w, http.StatusOK, response{Token: token})
}

func (apiCfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	bearerRefreshToken, err := auth.GetBearerToken(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid token", err)
		return
	}
	err = apiCfg.dbQueries.RevokeRefreshToken(r.Context(), bearerRefreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke refresh token", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
