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
	ExpiresInSeconds	*int		`json:"expires_in_seconds"`
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

	expiresInSeconds := 3600

	if params.ExpiresInSeconds != nil {
    	expiresInSeconds = min(*params.ExpiresInSeconds, 3600)
	}
	
	token, err := auth.MakeJWT(userInfo.ID, cfg.jwtSecret, time.Duration(expiresInSeconds)*time.Second)
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
