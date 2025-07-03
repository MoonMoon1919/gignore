package gignore

import (
	"errors"
	"io"
)

var (
	fileOpenError     = errors.New("error opening file")
	fileReadError     = errors.New("failed to read file content")
	fileCreationError = errors.New("failed to create file")
)

// LoadFile reads ignore file content from any io.Reader and populates the provided IgnoreFile.
// This function provides a flexible way to load ignore files from various sources including
// files, network connections, or in-memory buffers.
//
// Parameters:
//   - reader: Any io.Reader containing ignore file content to parse.
//   - ignoreFile: A pointer to the IgnoreFile instance to populate with the parsed rules.
//
// Returns an error if reading from the reader fails or if parsing the content fails.
//
// Example:
//
//	// Load from a file
//	file, err := os.Open(".gitignore")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	var ignoreFile IgnoreFile
//	err = LoadFile(file, &ignoreFile)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Load from a string
//	content := "*.log\nnode_modules/\n"
//	reader := strings.NewReader(content)
//	err = LoadFile(reader, &ignoreFile)
func LoadFile(reader io.Reader, ignoreFile *IgnoreFile) error {
	content, err := io.ReadAll(reader)
	if err != nil {
		return fileReadError
	}

	return Parse(string(content), ignoreFile)
}

// WriteFile writes an IgnoreFile to any io.Writer using the specified rendering options.
// This function provides a flexible way to output ignore files to various destinations
// including files, network connections, or in-memory buffers.
//
// Parameters:
//   - writer: Any io.Writer where the ignore file content should be written.
//   - ignoreFile: A pointer to the IgnoreFile instance to write.
//   - opts: The rendering options that control output formatting:
//     TrailingNewLine adds a newline character at the end if true.
//     HeaderComment adds a comment line at the top (prefixed with "# ") if non-empty.
//
// Returns an error if writing to the writer fails.
//
// Example:
//
//	// Write to a file
//	file, err := os.Create(".gitignore")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	opts := RenderOptions{TrailingNewLine: true}
//	err = WriteFile(file, &ignoreFile, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Write to stdout
//	err = WriteFile(os.Stdout, &ignoreFile, opts)
func WriteFile(writer io.Writer, ignoreFile *IgnoreFile, opts RenderOptions) error {
	content := Render(ignoreFile, opts)
	_, err := writer.Write([]byte(content))

	return err
}
