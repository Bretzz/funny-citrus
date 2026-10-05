package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"funny-citrus/v3/pkg/basic"
)

/*
## Endpoints
### Special Endpoints
- `GET /p`: resturns a JSON with a list of all projects. All project related actions will be under this route
- `GET /p/download?project=<project>`: downloads a project (zip'ed folder)
### Standard Endpoints
- `GET /`: index.html, all static assets under this route
- `GET /p/{project}`: returns a JSON:
```json
    {
        "resources": ["resource1.txt, resource2.png, ..."],     // array of strings
        "template": "#<project-name>\n..."                      // template.md as a string
    }
```
- `GET /p/{project}/{resource}` get the specific resource
*/

var PROJECT_DIR string = ""

var INVALID_PROJECT_NAMES = []string{"", ".", "/", "p"}

func BuildProjectPath(project string) (string, error) {
	project = filepath.Base(project)

	// fmt.Printf("Building path for project '%s'\n", project)

	// if project == "." || project == "/" || project == "p" || project == "create" || project == "delete" || project == "upload" {
	if basic.Contains(INVALID_PROJECT_NAMES, project) {
		return "<3", fmt.Errorf("Invalid project name")
	}

	projPath := filepath.Join(PROJECT_DIR, project)

	return projPath, nil
}

func main() {

	// Retreive data directory
	if len(os.Args) != 2 {
		fmt.Printf("Error, missing data directory\n")
		os.Exit(1)
	}
	PROJECT_DIR = os.Args[1]
	fmt.Printf("PROJECT_DIR: %s\n", os.Args[1])

	// Mux is needed to process dynamic URLs
	mux := http.NewServeMux()
	// Serve static content found in public/
	mux.Handle("/", http.FileServer(http.Dir("./public")))

	// GET endpoints
	mux.HandleFunc("GET /p", GetProjectList)
	mux.HandleFunc("GET /p/{project}", GetProjectMetadata)
	mux.HandleFunc("GET /p/{project}/{resource}", GetResource)
	// download
	mux.HandleFunc("GET /p/download", GetProjectDownload)

	// log.Fatal(http.ListenAndServe(":8080", logRequest(mux)))
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// Logging and enabling CORS
func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		enableCors(w)
		handler.ServeHTTP(w, r)
	})
}
