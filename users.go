package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aott33/chirpy/internal/auth"
	"github.com/aott33/chirpy/internal/database"
	"github.com/google/uuid"
)

type userParams struct {
	Email				string 		`json:"email"`
	Password			string		`json:"password"`
}

type tokenResponse struct {
	Token				string		`json:"token"`		
}

type User struct {
	ID        			uuid.UUID 	`json:"id"`
	CreatedAt 			time.Time 	`json:"created_at"`
	UpdatedAt 			time.Time 	`json:"updated_at"`
	Email    			string    	`json:"email"`
}

type UserLogin struct {
	ID				uuid.UUID 	`json:"id"`
	CreatedAt		time.Time 	`json:"created_at"`
	UpdatedAt		time.Time 	`json:"updated_at"`
	Email			string    	`json:"email"`
	Token			string		`json:"token"`
	RefreshToken	string		`json:"refresh_token"`
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	var params userParams

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}	
	
	userInfo, err := cfg.dbQueries.GetUserPassword(r.Context(), params.Email)
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, userInfo.HashedPassword)
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !match {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := auth.MakeJWT(userInfo.ID, cfg.jwtSecret) 
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	refreshToken, _ := auth.MakeRefreshToken()
	
	refreshTokenEntry, err := cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID: userInfo.ID,
		Token: refreshToken,
	})
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := UserLogin{
		ID: userInfo.ID,
		CreatedAt: userInfo.CreatedAt,
		UpdatedAt: userInfo.UpdatedAt,
		Email: userInfo.Email,
		Token: token,
		RefreshToken: refreshTokenEntry.Token,
	}

	dat, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)	

}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var params userParams

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userCreated, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email: params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := User{
		ID: userCreated.ID,
		CreatedAt: userCreated.CreatedAt,
		UpdatedAt: userCreated.UpdatedAt,
		Email: userCreated.Email,
	}

	dat, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(dat)	
}

func (cfg *apiConfig) updateUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	uuidVal, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {	
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusUnauthorized)	
		return
	}

	var params userParams

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userUpdated, err := cfg.dbQueries.UpdateUserInfo(r.Context(), database.UpdateUserInfoParams{
		Email: params.Email,
		HashedPassword: hashedPassword,
		ID: uuidVal,
	})
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := User{
		ID: userUpdated.ID,
		CreatedAt: userUpdated.CreatedAt,
		UpdatedAt: userUpdated.UpdatedAt,
		Email: userUpdated.Email,
	}

	dat, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)	
}

func (cfg *apiConfig) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {	
		writeJSON(w, http.StatusUnauthorized, errorResponse {
			Error: fmt.Sprintf("Something went wrong: %v", err),
		})	
		return
	}
	
	refreshTokenEntry, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {	
		writeJSON(w, http.StatusUnauthorized, errorResponse {
			Error: fmt.Sprintf("Something went wrong: %v", err),
		})	
		return
	}

	tokenRefreshed, err := auth.MakeJWT(refreshTokenEntry.UserID, cfg.jwtSecret)
	if err != nil {	
		writeJSON(w, http.StatusInternalServerError, errorResponse {
			Error: fmt.Sprintf("Something went wrong: %v", err),
		})	
		return
	}

	tokenResponse := tokenResponse {
		Token: tokenRefreshed,
	}

	dat, err := json.Marshal(tokenResponse)
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)	
}

func (cfg *apiConfig) revokeTokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {	
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	
	err = cfg.dbQueries.RevokeToken(r.Context(), token)
	if err != nil {	
		w.WriteHeader(http.StatusInternalServerError)	
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
