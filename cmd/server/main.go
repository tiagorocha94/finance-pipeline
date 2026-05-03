package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	dir := flag.String("dir", "./output", "directory to serve")
	port := flag.String("port", "8080", "port to listen on")
	flag.Parse()

	abs, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatalf("invalid dir: %v", err)
	}

	if _, err := os.Stat(abs); os.IsNotExist(err) {
		log.Fatalf("directory does not exist: %s", abs)
	}

	addr := "http://localhost:" + *port
	fmt.Printf("serving %s at %s\n", abs, addr)
	fmt.Println("press Ctrl+C to stop")

	http.Handle("/", http.FileServer(http.Dir(abs)))
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
