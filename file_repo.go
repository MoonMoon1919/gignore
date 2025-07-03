package gignore

import "os"

type FileRepository struct {
	renderOptions RenderOptions
}

// NewFileRepository creates a new FileRepository with the specified rendering options.
// The FileRepository provides file-based persistence operations for IgnoreFile instances.
//
// Parameters:
//   - opts: The rendering options that control how IgnoreFiles are formatted when saved.
//
// Returns a FileRepository configured with the provided options.
//
// Example:
//
//	opts := RenderOptions{
//	    TrailingNewLine: true,
//	    HeaderComment:   "Generated ignore file - do not edit manually",
//	}
//	repo := NewFileRepository(opts)
//	err := repo.Save(".gitignore", ignoreFile)
func NewFileRepository(opts RenderOptions) FileRepository {
	return FileRepository{
		renderOptions: opts,
	}
}

// Load reads an ignore file from the specified path and populates the provided IgnoreFile.
// The file is automatically opened and closed during the operation.
//
// Parameters:
//   - path: The file system path to the ignore file to load.
//   - ignoreFile: A pointer to the IgnoreFile instance to populate with the loaded rules.
//
// Returns an error if the file cannot be opened or if parsing fails.
//
// Example:
//
//	var ignoreFile IgnoreFile
//	err := repo.Load(".gitignore", &ignoreFile)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (f FileRepository) Load(path string, ignoreFile *IgnoreFile) error {
	file, err := os.Open(path)
	if err != nil {
		return fileOpenError
	}

	defer file.Close()

	return LoadFile(file, ignoreFile)
}

// Save writes an IgnoreFile to the specified path using the repository's rendering options.
// The file is automatically created (or overwritten if it exists) and closed during the operation.
// The output format is controlled by the RenderOptions specified when creating the repository.
//
// Parameters:
//   - path: The file system path where the ignore file should be saved.
//   - ignoreFile: A pointer to the IgnoreFile instance to save.
//
// Returns an error if the file cannot be created or if writing fails.
//
// Example:
//
//	opts := RenderOptions{
//	    TrailingNewLine: true,
//	    HeaderComment:   "# Auto-generated .gitignore",
//	}
//	repo := NewFileRepository(opts)
//	err := repo.Save(".gitignore", &ignoreFile)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (f FileRepository) Save(path string, ignoreFile *IgnoreFile) error {
	file, err := os.Create(path)
	if err != nil {
		return fileCreationError
	}

	defer file.Close()

	return WriteFile(file, ignoreFile, f.renderOptions)
}
