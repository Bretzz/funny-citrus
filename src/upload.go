package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func isValidFileType(file []byte) bool {
	fileType := http.DetectContentType(file)
	return strings.HasPrefix(fileType, "image/") // Only allow images
}

func sanitizeName(name string) (string, error) {
	safeName := filepath.Base(name)

	if safeName == "." || safeName == "/" {
		return "<3", fmt.Errorf("invalid path parameters")
	}

	return safeName, nil
}

func createFileInProject(project string, filename string) (*os.File, error) {
	// Sanitize project and file
	safeProject, err := sanitizeName(project)
	if err != nil {
		return nil, err
	}
	safeFilename, err := sanitizeName(filename)
	if err != nil {
		return nil, err
	}

	// Form the complete path
	projectDir := filepath.Join(UPLOAD_FOLDER, safeProject)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("Project doesn't exist")
	}
	fullPath := filepath.Join(projectDir, safeFilename)
	if _, err := os.Stat(fullPath); err == nil {
		return nil, fmt.Errorf("Resource already exists")
	}

	// Create and return the destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}

	return dst, nil
}

func multiUpload(w http.ResponseWriter, r *http.Request, project, filename string) {
	r.ParseMultipartForm(10 << 20) // 10MB limit

	file, _, err := r.FormFile("myFile")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 1. Read first 512 bytes to inspect content type
	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil && err != io.EOF {
		http.Error(w, "Error reading file header", http.StatusInternalServerError)
		return
	}

	// 2. Validate file type BEFORE creating destination file
	if !isValidFileType(buff) {
		http.Error(w, "Invalid file type: must be an image", http.StatusUnsupportedMediaType)
		return
	}

	// 3. Reset file read pointer back to the beginning
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Error processing file", http.StatusInternalServerError)
		return
	}

	// 4. Create destination file inside project folder and stream
	dst, err := createFileInProject(project, filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Successfully uploaded %s to project '%s'\n", filename, project)
}

func rawUpload(w http.ResponseWriter, r *http.Request, project, filename string) {
	// Restrict max request payload size (10MB)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	// Read the first 512 bytes for type validation
	buff := make([]byte, 512)
	n, err := io.ReadFull(r.Body, buff)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	buff = buff[:n]

	if !isValidFileType(buff) {
		http.Error(w, "Invalid file type: must be an image", http.StatusUnsupportedMediaType)
		return
	}

	dst, err := createFileInProject(project, filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Write the 512 bytes already read to disk
	if _, err := dst.Write(buff); err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	// Stream remaining body content to disk
	if _, err := io.Copy(dst, r.Body); err != nil {
		http.Error(w, "Error saving rest of file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Successfully saved %s to project '%s'\n", filename, project)
}

func fileUploadHandler(w http.ResponseWriter, r *http.Request) {
	// Extract route wildcards (Go 1.22+)
	project := r.PathValue("project")
	filename := r.PathValue("filename")

	if project == "" {
		http.Error(w, "Project path parameter required", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		multiUpload(w, r, project, filename)
	} else {
		rawUpload(w, r, project, filename)
	}
}
