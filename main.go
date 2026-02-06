package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	mu sync.RWMutex
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./static"
	}

	_, err := os.Stat(staticDir)
	if err != nil {
		log.Fatalf("ERROR: STATIC_DIR='%s' does not exist", staticDir)
	}

	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" {
		dataFile = "pad.txt"
	}

	log.Printf("INFO: Starting server on port %s", port)
	log.Printf("INFO: Serving static files from: %s", staticDir)
	log.Printf("INFO: Data file: %s", dataFile)

	http.Handle("/", http.FileServer(http.Dir(staticDir)))
	http.HandleFunc("/api/text", func(w http.ResponseWriter, r *http.Request) {
		handleText(w, r, dataFile)
	})

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}

func handleText(w http.ResponseWriter, r *http.Request, dataFile string) {
	mu.Lock()
	defer mu.Unlock()

	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := os.WriteFile(dataFile, body, 0600); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	content, err := os.ReadFile(dataFile)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(content); err != nil {
		return
	}
}
