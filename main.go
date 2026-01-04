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
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Error opening database: %v", err)
		return
	}

	dbQueries := database.New(db)

	mux := http.NewServeMux()

	apiCfg := &apiConfig{}
	apiCfg.dbQueries = *dbQueries

	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServerHandler))

	mux.HandleFunc("GET /api/healthz", healthHandler)

	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)

	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

	server := &http.Server{Addr: ":8080", Handler: mux}

	fmt.Printf("Server starting on port 8080\n")

	server.ListenAndServe()
}
