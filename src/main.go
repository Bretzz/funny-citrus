package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const PORT = ":8080"
const UPLOAD_FOLDER string = "data"

/* Mutex logic in GO */
/* var counter int
var mutex = &sync.Mutex{}

func incrementCounter(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	counter++
	fmt.Fprintf(w, strconv.Itoa(counter))
	mutex.Unlock()
} */

func createProjectHandler(w http.ResponseWriter, r *http.Request) {
	// Extract 'projectName' from the URL wildcard
	projectName := r.PathValue("project")

	// Sanitize path to prevent directory traversal (e.g., ../ attacks)
	safeProjectName := filepath.Base(projectName)
	if safeProjectName == "." || safeProjectName == "/" {
		http.Error(w, "Invalid project name", http.StatusBadRequest)
		return
	}

	// Build directory path: uploads/<projectName>
	projectDir := filepath.Join(UPLOAD_FOLDER, safeProjectName)
	if _, err := os.Stat(projectDir); err == nil {
		http.Error(w, "Project already exists", http.StatusBadRequest)
		return
	}

	// Create directory (0755: read/write/execute for owner, read/execute for others)
	err := os.MkdirAll(projectDir, 0755)
	if err != nil {
		http.Error(w, "Failed to create project folder", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Project folder '%s' created successfully at %s\n", safeProjectName, projectDir)
}

func main() {
	fmt.Printf("Hello World\n")

	// Initialize a server multiplexer to parse complex requests
	mux := http.NewServeMux()
	// Set static serving directory
	fileServer := http.FileServer(http.Dir("./static"))

	// Serve static assets at root (e.g. GET /index.html -> ./static/index.html)
	mux.Handle("GET /", fileServer)

	// Handles Project creation and POST of files
	mux.HandleFunc("POST /{project}", createProjectHandler)
	mux.HandleFunc("POST /{project}/{filename}", fileUploadHandler)

	mux.HandleFunc("GET /hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})

	log.Fatal(http.ListenAndServe(":8080", mux))

}
