package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
)

func getProjectHandler(w http.ResponseWriter, r *http.Request) {
	// Clean the project parameter to strip accidental trailing slashes
	project := strings.Trim(r.PathValue("project"), "/")
	project = filepath.Base(project)

	if project == "." || project == "" {
		http.Error(w, "Invalid project name", http.StatusBadRequest)
		return
	}

	// 1. Verify project directory exists
	projectDir := filepath.Join(UPLOAD_DIR, project)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// 2. Read template.md from project directory
	templateBytes, err := os.ReadFile(filepath.Join(projectDir, PROJECT_TEMPLATE))
	if err != nil {
		// Read template.md from root
		templateBytes, err = os.ReadFile(PROJECT_TEMPLATE)
		if err != nil {
			http.Error(w, "Failed to read template file", http.StatusInternalServerError)
			return
		}
	}
	content := string(templateBytes)

	// 2. Replace <project-name> placeholder
	content = strings.ReplaceAll(content, "<project-name>", project)

	// 3. Construct the download link
	downloadURL := fmt.Sprintf("/%s/download", project)
	downloadLinkMarkdown := fmt.Sprintf("[download_project](%s)\n\n", downloadURL)

	// 4. Replace [download_project] directly where it appears in the template
	if strings.Contains(content, "[download_project]") {
		content = strings.Replace(content, "[download_project]", downloadLinkMarkdown, 1)
	} else {
		// Fallback: If [download_project] is missing from template.md, insert below header
		projectHeader := fmt.Sprintf("# %s\n\n%s", project, downloadLinkMarkdown)
		content = strings.Replace(content, fmt.Sprintf("# %s", project), projectHeader, 1)
	}

	// 5. Scan project directory for files and build resource list
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	var resourcesList strings.Builder
	for _, entry := range entries {
		if !entry.IsDir() {
			filename := entry.Name()
			ext := strings.ToLower(filepath.Ext(filename))

			// Escape spaces and special characters for URLs (e.g. "my photo.png" -> "my%20photo.png")
			escapedFilename := url.PathEscape(filename)
			resourceURL := fmt.Sprintf("/%s/%s", project, escapedFilename)

			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp" {
				// Clickable header + image preview
				resourcesList.WriteString(fmt.Sprintf("\n### [%s](%s)\n![%s](%s)\n", filename, resourceURL, filename, resourceURL))
			} else {
				// Non-image resources render as links
				resourcesList.WriteString(fmt.Sprintf("\n### [%s](%s)\n", filename, resourceURL))
			}
		}
	}

	if resourcesList.Len() == 0 {
		resourcesList.WriteString("\n*No resources uploaded yet.*")
	}

	// 6. Inject resource list into the markdown content
	targetHeader := "## Resources"
	if strings.Contains(content, targetHeader) {
		content = strings.Replace(content, targetHeader, targetHeader+"\n"+resourcesList.String(), 1)
	} else {
		content += "\n\n" + targetHeader + "\n" + resourcesList.String()
	}

	// 7. Convert Markdown to HTML
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(content), &buf); err != nil {
		http.Error(w, "Failed to parse markdown", http.StatusInternalServerError)
		return
	}

	// 8. Wrap in HTML scaffold with restricted image preview sizing
	htmlOutput := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>%s</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; 
            max-width: 800px; 
            margin: 40px auto; 
            padding: 0 20px; 
            line-height: 1.6; 
            color: #24292e;
        }
        /* Restricts preview size for all embedded markdown images */
        img { 
            max-width: 280px; 
            max-height: 200px; 
            width: auto; 
            height: auto; 
            object-fit: contain; 
            border-radius: 6px; 
            border: 1px solid #d0d7de; 
            margin-top: 8px; 
            display: block; 
        }
        h3 { 
            margin-top: 20px; 
            margin-bottom: 4px; 
            font-size: 1.1rem;
        }
        a { color: #0969da; text-decoration: none; }
        a:hover { text-decoration: underline; }
		.grid {
			display: flex;
			flex-direction: column;
			gap: 8px;
			width: 100%%;
			max-width: 500px; /* Adjust to match your content width */
		}

		.grid form {
			display: flex;
			gap: 8px;
			width: 100%%;
		}

		.grid form input[type="file"] {
			flex: 1; /* Pushes the button neatly to the right */
			box-sizing: border-box;
		}

		.upload-status {
			width: 100%%;
			height: 40px;
			border: 1px solid #d0d7de;
			border-radius: 6px;
			background-color: #f6f8fa;
			box-sizing: border-box;
		}
    </style>
</head>
<body>
	<div class="grid">
		<form action="/%s/upload" method="POST" enctype="multipart/form-data" target="upload_frame">
			<input type="file" name="file" multiple required>
			<button type="submit">Upload</button>
		</form>

		<!-- Visible target frame that shows server status or http.Error messages -->
		<iframe name="upload_frame" class="upload-status"></iframe>
	</div>
    %s
</body>
</html>`, project, project, buf.String())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlOutput))
}

func downloadProjectHandler(w http.ResponseWriter, r *http.Request) {
	project := strings.Trim(r.PathValue("project"), "/")
	project = filepath.Base(project)

	if project == "." || project == "" {
		http.Error(w, "Invalid project name", http.StatusBadRequest)
		return
	}

	projectDir := filepath.Join(UPLOAD_DIR, project)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Set headers to prompt a file download in the browser
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", project))

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	// Walk through the project directory and write files into the zip archive
	err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories themselves (files inside them will be added)
		if info.IsDir() {
			return nil
		}

		// Calculate relative path inside the zip file
		relPath, err := filepath.Rel(projectDir, path)
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
	}
}

func getResourceHandler(w http.ResponseWriter, r *http.Request) {
	project := filepath.Base(r.PathValue("project"))
	resource := r.PathValue("resource")

	// If resource is empty (e.g. GET /demo/), delegate back to getProjectHandler
	if strings.TrimSpace(resource) == "" || resource == "/" {
		getProjectHandler(w, r)
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

	http.ServeFile(w, r, filePath)
}

// IndexData holds the slice of project folder names for the HTML template
type IndexData struct {
	Projects []string
}

func listProjectsHandler(w http.ResponseWriter, r *http.Request) {
	entries, _ := os.ReadDir(UPLOAD_DIR)
	var projects []string
	for _, entry := range entries {
		if entry.IsDir() {
			projects = append(projects, entry.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}
