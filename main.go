package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

//go:embed static
var embedFS embed.FS

var (
	mu sync.RWMutex

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	clients     = make(map[*websocket.Conn]bool)
	currentText []byte
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	staticFiles, err := fs.Sub(embedFS, "static")
	if err != nil {
		log.Fatal(err)
	}

	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" {
		dataFile = "pad.txt"
	}

	content, err := os.ReadFile(dataFile)
	if err == nil {
		currentText = content
	}

	log.Printf("INFO: Starting server on port %s", port)
	log.Printf("INFO: Data file: %s", dataFile)

	http.Handle("/", http.FileServer(http.FS(staticFiles)))
	http.HandleFunc("/api/text", handleText)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWS(w, r, dataFile)
	})

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}

func handleText(w http.ResponseWriter, _ *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	if _, err := w.Write(currentText); err != nil {
		log.Println("write response:", err)
	}
}

func handleWS(w http.ResponseWriter, r *http.Request, dataFile string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	mu.Lock()
	clients[conn] = true
	if err := conn.WriteMessage(websocket.TextMessage, currentText); err != nil {
		delete(clients, conn)
		mu.Unlock()
		_ = conn.Close()
		return
	}
	mu.Unlock()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		mu.Lock()
		currentText = message

		if err := os.WriteFile(dataFile, message, 0600); err != nil {
			log.Println("write file:", err)
		}

		for c := range clients {
			if c == conn {
				continue
			}
			if err := c.WriteMessage(websocket.TextMessage, message); err != nil {
				_ = c.Close()
				delete(clients, c)
			}
		}
		mu.Unlock()
	}

	mu.Lock()
	delete(clients, conn)
	mu.Unlock()

	if err := conn.Close(); err != nil {
		log.Println("close:", err)
	}
}
