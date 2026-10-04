package pfile

import (
	"fmt"
	"os"
	"path/filepath"
)

// Returns a set of all subdirs of a specific directory.
// Returns an error if it's an absolute path.
func ListSubdirs(parent string) ([]string, error) {

	parent = filepath.Clean(parent)
	if !filepath.IsLocal(parent) {
		return nil, fmt.Errorf("Error: The provided path isn't local")
	}

	entries, err := os.ReadDir(parent)
	if err != nil {
		return nil, fmt.Errorf("Error while opening directory")
	}

	var subs []string
	for _, entry := range entries {
		if entry.IsDir() {
			subs = append(subs, entry.Name())
		}
	}

	return subs, nil
}

// Returns a set of all files of a specific directory.
// Returns an error if it's an absolute path.
func ListFiles(dir string) ([]string, error) {

	// dir = filepath.Clean(dir)
	if !filepath.IsLocal(dir) {
		return nil, fmt.Errorf("Error: The provided path isn't local")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("Error while opening directory")
	}

	var subs []string
	for _, entry := range entries {
		fmt.Printf("Found '%s' in '%s'\n", entry.Name(), dir)
		if !entry.IsDir() {
			subs = append(subs, entry.Name())
		}
	}

	return subs, nil
}

// Assuming the folder is ok, here only check if is local
func CreateFileInFolder(path string, filename string) (*os.File, error) {

	if !filepath.IsLocal(path) {
		return nil, fmt.Errorf("Not a local path")
	}

	// Form the complete path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("Project not found")
	}

	// Prevent directory traversal
	filename = filepath.Base(filename)
	filePath := filepath.Join(path, filename)

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
