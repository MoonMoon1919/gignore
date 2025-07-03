package gignore

import "strings"

type RenderOptions struct {
	TrailingNewLine bool
	HeaderComment   string
}

// Render converts an IgnoreFile to its string representation using the specified formatting options.
// The function generates a properly formatted ignore file content that can be written to disk or
// used for display purposes.
//
// Parameters:
//   - ignoreFile: A pointer to the IgnoreFile instance to render.
//   - options: The rendering options that control output formatting:
//     TrailingNewLine adds a newline character at the end of the output if true.
//     HeaderComment adds a comment line at the top of the file (prefixed with "# ")
//     followed by a blank line if the comment is non-empty.
//
// Returns a string containing the formatted ignore file content. Each rule appears on
// its own line, with rules rendered in their current order within the IgnoreFile.
//
// Example:
//
//	opts := RenderOptions{
//	    TrailingNewLine: true,
//	    HeaderComment:   "Generated ignore file - do not edit manually",
//	}
//	content := Render(&ignoreFile, opts)
//	fmt.Print(content)
//	// Output:
//	// # Generated ignore file - do not edit manually
//	//
//	// *.log
//	// node_modules/
//	// build/
func Render(ignoreFile *IgnoreFile, options RenderOptions) string {
	var lines []string

	if len(options.HeaderComment) > 0 {
		lines = append(lines, "# "+options.HeaderComment)
		lines = append(lines, "") // blank line after header comment
	}

	for _, rule := range ignoreFile.Rules() {
		lines = append(lines, rule.Render())
	}

	result := strings.Join(lines, "\n")
	if options.TrailingNewLine {
		result += "\n"
	}

	return result
}
