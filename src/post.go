package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func isValidFileType(file []byte) bool {
	fileType := http.DetectContentType(file)
	return strings.HasPrefix(fileType, "image/") // Only allow images
}

func createFileInProject(project string, resource string) (*os.File, error) {
	// Clean the project parameter to strip accidental trailing slashes
	project = filepath.Base(project)

	if project == "." || project == "" || project == "download" {
		return nil, fmt.Errorf("Invalid project name")
	}

	// Form the complete path
	projectDir := filepath.Join(UPLOAD_DIR, project)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("Project not found")
	}

	// Prevent directory traversal
	resource = filepath.Base(resource)
	filePath := filepath.Join(UPLOAD_DIR, project, resource)

	fmt.Printf("checking if %s exists\n", filePath)

	if _, err := os.Stat(filePath); err == nil {
		return nil, fmt.Errorf("Resource already exists")
	}

	// Create and return the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	return dst, nil
}

func multiUpload(w http.ResponseWriter, r *http.Request, project, resource string) bool {
	// 1. Parse multipart form with a memory limit (e.g., 32 MB)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return false
	}

	// 2. Extract files from multipart header ("files" or "file")
	files := r.MultipartForm.File["file"] // Form input name="file"
	if len(files) == 0 {
		http.Error(w, "No files provided", http.StatusBadRequest)
		return false
	}

	var savedFiles []string

	// 3. Iterate and save each file
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}
		defer file.Close()

		// 4. Create destination file inside project folder and stream
		dst, err := createFileInProject(project, fileHeader.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return false
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Failed to copy file content", http.StatusInternalServerError)
			return false
		}

		savedFiles = append(savedFiles, fileHeader.Filename)
	}

	// w.WriteHeader(http.StatusOK)
	// fmt.Fprintf(w, "Successfully uploaded %d files: %s\n", len(savedFiles), strings.Join(savedFiles, ", "))
	return true
}

func rawUpload(w http.ResponseWriter, r *http.Request, project, resource string) bool {
	// Restrict max request payload size (10MB)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	// Read the first 512 bytes for type validation
	buff := make([]byte, 512)
	n, err := io.ReadFull(r.Body, buff)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return false
	}

	buff = buff[:n]

	if !isValidFileType(buff) {
		http.Error(w, "Invalid file type: must be an image", http.StatusUnsupportedMediaType)
		return false
	}

	dst, err := createFileInProject(project, resource)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	defer dst.Close()

	// Write the 512 bytes already read to disk
	if _, err := dst.Write(buff); err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return false
	}

	// Stream remaining body content to disk
	if _, err := io.Copy(dst, r.Body); err != nil {
		http.Error(w, "Error saving rest of file", http.StatusInternalServerError)
		return false
	}

	// fmt.Fprintf(w, "Successfully saved %s to project '%s'\n", resource, project)
	return true
}

func postResourceHandler(w http.ResponseWriter, r *http.Request) {
	// Extract route wildcards (Go 1.22+)
	project := strings.Trim(r.PathValue("project"), "/")
	resource := r.PathValue("resource")

	if project == "" {
		http.Error(w, "Project path parameter required", http.StatusBadRequest)
		return
	}

	if project == "download" || project == "create" {
		http.Error(w, "Invalid project name", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")

	var ok bool
	if strings.HasPrefix(contentType, "multipart/form-data") {
		ok = multiUpload(w, r, project, resource)
	} else {
		ok = rawUpload(w, r, project, resource)
	}

	// Redirect back to the exact same path the form was submitted to
	// (This causes the browser to issue a GET request and reload the page)
	if ok == true {
		projectURL := path.Dir(r.URL.Path)
		http.Redirect(w, r, projectURL, http.StatusSeeOther)
	}
}

// Creates a project folder inside the UPLOAD_DIR
func postProjectHandler(w http.ResponseWriter, r *http.Request) {
	// Extract 'projectName' from the URL wildcard
	project := r.FormValue("project")
	if project == "" {
		project = r.PathValue("project")
	}

	// Sanitize path to prevent directory traversal (e.g., ../ attacks)
	project = filepath.Base(project)
	if project == "." || project == "/" || project == "projects" {
		http.Error(w, "Invalid project name", http.StatusBadRequest)
		return
	}

	// Build directory path: uploads/<projectName>
	projectDir := filepath.Join(UPLOAD_DIR, project)
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

	// 1. Read the template file from the root directory
	templateData, err := os.ReadFile(PROJECT_TEMPLATE)
	if err != nil {
		http.Error(w, "Failed to read template file", http.StatusInternalServerError)
		return
	}

	// 2. Define the destination path: uploads/<safeProjectName>/template.md
	destPath := filepath.Join(projectDir, PROJECT_TEMPLATE)

	// 3. Write the template contents to the new project folder (0644 permissions)
	err = os.WriteFile(destPath, templateData, 0644)
	if err != nil {
		http.Error(w, "Failed to copy template to project folder", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Project folder '%s' created successfully at %s\n", project, projectDir)
}
