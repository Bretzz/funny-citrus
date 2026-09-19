package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func deleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	// Clean the project parameter to strip accidental trailing slashes
	project := strings.Trim(r.PathValue("project"), "/")
	project = filepath.Base(project)

	if project == "." || project == "" {
		http.Error(w, "Invalid project name", http.StatusBadRequest)
		return
	}

	// Form the complete path
	projectDir := filepath.Join(UPLOAD_DIR, project)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		http.Error(w, "Project not found", http.StatusBadRequest)
		return
	}

	err := os.RemoveAll(projectDir)
	if err != nil {
		http.Error(w, "Error deleting the project", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Project deleted successfully"))
}

func deleteResourceHandler(w http.ResponseWriter, r *http.Request) {
	project := filepath.Base(r.PathValue("project"))
	resource := r.PathValue("resource")

	// If resource is empty (e.g. GET /demo/), delegate back to getProjectHandler
	if strings.TrimSpace(resource) == "" || resource == "/" {
		http.Error(w, "Invalid resource name", http.StatusBadRequest)
		return
	}

	// Prevent directory traversal
	resource = filepath.Base(resource)
	filePath := filepath.Join(UPLOAD_DIR, project, resource)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	}

	// Delete the resource
	if err := os.Remove(filePath); err != nil {
		http.Error(w, "Error deleting the resource", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Resource deleted successfully"))
}
