package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	abs, err := filepath.Abs("./output")
	if err != nil {
		log.Fatalf("invalid dir: %v", err)
	}

	if _, err := os.Stat(abs); os.IsNotExist(err) {
		log.Fatalf("directory does not exist: %s", abs)
	}

	addr := "http://localhost:8080"
	fmt.Printf("serving %s at %s\n", abs, addr)
	fmt.Println("press Ctrl+C to stop")

	http.Handle("/", http.FileServer(http.Dir(abs)))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
