package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/aott33/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits	atomic.Int32
	dbQueries		database.Queries
	platform		string
}

func setupRoutes(mux *http.ServeMux, cfg *apiConfig) {
	// Static Files
	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", cfg.middlewareMetricsInc(fileServerHandler))

	// Admin Routes
	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", cfg.resetHandler)
	
	// API Routes
	mux.HandleFunc("POST /api/chirps", validateChirpHandler)
	mux.HandleFunc("POST /api/users", cfg.createUserHandler)
	mux.HandleFunc("GET /api/healthz", healthHandler)


}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Error opening database: %v", err)
		return
	}

	dbQueries := database.New(db)

	mux := http.NewServeMux()
	apiCfg := &apiConfig{
		dbQueries: *dbQueries,
		platform: platform,	
	}
	
	setupRoutes(mux, apiCfg)	

	server := &http.Server{Addr: ":8080", Handler: mux}
	fmt.Printf("Server starting on port 8080\n")
	server.ListenAndServe()
}
