package main

import (
	"fmt"
	"log"
	"net/http"
)

const PORT = ":8080"
const UPLOAD_DIR string = "data"
const PROJECT_TEMPLATE string = "template.md"

/* Mutex logic in GO */
/* var counter int
var mutex = &sync.Mutex{}

func incrementCounter(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	counter++
	fmt.Fprintf(w, strconv.Itoa(counter))
	mutex.Unlock()
} */

func main() {
	fmt.Println("Server starting on :8080...")

	mux := http.NewServeMux()

	// Rule 6: Serve root "/" (Serves ./static/index.html, handles /login redirect flow)
	// staticServer := http.FileServer(http.Dir("./static"))
	mux.Handle("GET /", http.FileServer(http.Dir("./static")))

	// Auth endpoint referenced in Rule 6
	// mux.HandleFunc("GET /login", loginHandler)

	// List projects at the root URL
	mux.HandleFunc("GET /projects", listProjectsHandler)

	// Download the whole project as a zip
	mux.HandleFunc("GET /{project}/download", downloadProjectHandler)

	// Render project summary page
	mux.HandleFunc("GET /{project}/{$}", getProjectHandler)
	mux.HandleFunc("GET /{project}", getProjectHandler)

	// Project & Resource CRUD
	mux.HandleFunc("POST /{project}", postProjectHandler)
	mux.HandleFunc("DELETE /{project}", deleteProjectHandler)
	mux.HandleFunc("GET /{project}/{resource}", getResourceHandler)
	mux.HandleFunc("POST /{project}/{resource}", postResourceHandler)
	mux.HandleFunc("DELETE /{project}/{resource}", deleteResourceHandler)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
