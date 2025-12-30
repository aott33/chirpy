package main

import (
	"net/http"
	"sync/atomic"
	"fmt"
)

type apiConfig struct {
	fileserverHits	atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hits: %d", cfg.fileserverHits.Load())
}

func (cfg *apiConfig) resetHanlder(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}


func main() {
	mux := http.NewServeMux()

	apiCfg := &apiConfig{}

	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServerHandler))

	mux.HandleFunc("GET /healthz", healthHandler)

	mux.HandleFunc("GET /metrics", apiCfg.metricsHandler)

	mux.HandleFunc("POST /reset", apiCfg.resetHanlder)

	server := &http.Server{Addr: ":8080", Handler: mux}

	server.ListenAndServe()
}
