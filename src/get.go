package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"funny-citrus/v3/pkg/pfile"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func GetProjectList(w http.ResponseWriter, r *http.Request) {
	projects, err := pfile.ListSubdirs(PROJECT_DIR)

	// #debug
	// fmt.Printf("dir %s\n", PROJECT_DIR)

	if err != nil {
		http.Error(w, "Unable to open project directory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

type ProjectResource struct {
	Filename     string `json:"filename"`
	TimeOfUpload string `json:"time_of_upload"`
	Size         int64  `json:"size"`
}

// ListResources returns a list of all files in a specific directory.
// Returns an error if the path is not local (e.g., absolute or escapes working dir).
func ListResources(dir string) ([]ProjectResource, error) {

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var resources []ProjectResource
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			// Skip files where metadata cannot be retrieved (or handle error as needed)
			res := ProjectResource{
				Filename:     entry.Name(),
				TimeOfUpload: "Error",
				Size:         0,
			}
			resources = append(resources, res)
			continue
		}

		res := ProjectResource{
			Filename:     entry.Name(),
			TimeOfUpload: info.ModTime().Format(time.RFC3339),
			Size:         info.Size(),
		}
		resources = append(resources, res)
	}

	return resources, nil
}

type ProjectMetadata struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Resources   []ProjectResource `json:"resources"`
	ResCount    int               `json:"res_count"`
	ProjectSize int64             `json:"project_size"`
}

// Expected a valid path, returns the project metadata
func BuildProjectMetadata(path string) (ProjectMetadata, error) {

	var metadata ProjectMetadata

	// Get Title & description
	metadata.Title = filepath.Base(path)
	metadata.Description = "interesting project"

	// Get resources
	resources, err := ListResources(path)
	if err != nil {
		return metadata, err
	}
	metadata.Resources = resources
	metadata.ResCount = len(resources)

	// Calculate project size
	var size int64 = 0
	for _, res := range resources {
		size += res.Size
	}
	metadata.ProjectSize = size

	return metadata, nil
}

func GetProjectMetadata(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")

	// Build project path
	projectPath, err := BuildProjectPath(project)
	if err != nil {
		http.Error(w, "Invalid project name", http.StatusBadRequest)
		return
	}

	// Read from project directory
	metadata, err := BuildProjectMetadata(projectPath)
	if err != nil {
		http.Error(w, "Unable to parse metadata", http.StatusInternalServerError)
		return
	}

	// Print the JSON object #debug
	// fmt.Print("This is what I ret\n", metadata)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

func GetResource(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")
	resource := filepath.Base(r.PathValue("resource"))

	// Build project path
	projectPath, err := BuildProjectPath(project)
	if err != nil {
		http.Error(w, "Invalid project name", http.StatusBadRequest)
		return
	}

	if info, err := os.Stat(projectPath); err != nil || !info.IsDir() {
		http.Error(w, "Project not found", http.StatusBadRequest)
		return
	}

	resPath := filepath.Join(projectPath, resource)
	if _, err := os.Stat(resPath); os.IsNotExist(err) {
		http.Error(w, "Resource not found", http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, resPath)
}

func GetProjectDownload(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")

	if project == "" {
		http.Error(w, "Missing query parameter: project", http.StatusBadRequest)
		return
	}

	// Build project path
	projectPath, err := BuildProjectPath(project)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		http.Error(w, "Project not found", http.StatusBadRequest)
		return
	}

	// Set headers to prompt a file download in the browser
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", project))

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	// Walk through the project directory and write files into the zip archive
	err = filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories themselves (files inside them will be added)
		if info.IsDir() {
			return nil
		}

		// Calculate relative path inside the zip file
		relPath, err := filepath.Rel(projectPath, path)
		if err != nil {
			return err
		}

		// Create file entry inside zip
		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		fileToZip, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fileToZip.Close()

		_, err = io.Copy(writer, fileToZip)
		return err
	})

	if err != nil {
		// Log error if streaming fails mid-way
		fmt.Printf("Error zipping project %s: %v\n", project, err)
		http.Error(w, "Error zipping project", http.StatusInternalServerError)
	}
}
