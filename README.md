# GIGNORE

A Go library for programmatically managing `.gitignore` and `.dockignore` files with intelligent conflict detection and automatic resolution.

## Features

- 🔍 **Intelligent conflict detection** - Finds redundant, unreachable, and ineffective rules
- 🔧 **Automatic conflict resolution** - Smart reordering and cleanup of ignore files
- 📁 **Repository abstraction** - Works with files, memory, or custom storage backends
- ✅ **Comprehensive validation** - Prevents invalid patterns and configurations
- 🎯 **Type-safe rule creation** - File, directory, extension, and glob patterns

## Installation

```bash
go get github.com/MoonMoon1919/gignore-cli
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/MoonMoon1919/gignore-cli"
)

func main() {
    // Create a new ignore file service
    repo := gignore.NewFileRepository(gignore.RenderOptions{})
    service := gignore.NewService(repo)

    // Add rules to .gitignore
    service.AddExtensionRule(".gitignore", "log", gignore.INCLUDE)
    service.AddDirectoryRule(".gitignore", "build", gignore.RECURSIVE, gignore.INCLUDE)
    service.AddFileRule(".gitignore", "config.json", gignore.INCLUDE)

    // Create exceptions
    service.AddFileRule(".gitignore", "build/important.txt", gignore.EXCLUDE)

    // Automatically fix any conflicts
    fixes, err := service.AutoFix(".gitignore", 10)
    if err != nil {
        panic(err)
    }

    for _, fix := range fixes {
        fmt.Println("Applied:", fix)
    }
}
```

## Core Concepts

### Rule Types

**File Rules** - Exact file paths
```go
service.AddFileRule(path, "src/main.go", gignore.INCLUDE)
// Generates: src/main.go
```

**Extension Rules** - File extensions
```go
service.AddExtensionRule(path, "log", gignore.INCLUDE)
// Generates: *.log
```

**Directory Rules** - Directory patterns with different modes
```go
// Just the directory
service.AddDirectoryRule(path, "build", gignore.DIRECTORY, gignore.INCLUDE)
// Generates: build/

// Everything in directory recursively
service.AddDirectoryRule(path, "build", gignore.RECURSIVE, gignore.INCLUDE)
// Generates: build/**

// Direct children only
service.AddDirectoryRule(path, "build", gignore.CHILDREN, gignore.INCLUDE)
// Generates: build/*

// Directory anywhere in tree
service.AddDirectoryRule(path, "temp", gignore.ANYWHERE, gignore.INCLUDE)
// Generates: **/temp

// Root-level only (leading slash)
service.AddDirectoryRule(path, "node_modules", gignore.ROOT_ONLY, gignore.INCLUDE)
// Generates: /node_modules
```

**Glob Rules** - Complex patterns
```go
service.AddGlobRule(path, "temp*.log", gignore.INCLUDE)
// Generates: temp*.log
```

### Actions

- `gignore.INCLUDE` - Ignore this pattern (no `!` prefix)
- `gignore.EXCLUDE` - Don't ignore this pattern (`!` prefix for exceptions)

### Conflict Detection

The library automatically detects three types of conflicts:

**Semantic Conflicts** - Same pattern with opposite actions
```gitignore
config.json
!config.json  # ← Conflicting actions
```

**Redundant Rules** - Duplicate patterns
```gitignore
*.log
*.log  # ← Redundant
```

**Unreachable Rules** - Broader patterns make specific ones meaningless
```gitignore
build/**
build/  # ← Unreachable (build/** already covers this)
```

**Ineffective Rules** - Exception rules with no prior exclusion
```gitignore
!build/important.txt  # ← Ineffective (nothing to override)
build/**
```

## Advanced Usage

### Manual Conflict Analysis

```go
// Load existing .gitignore
var ignoreFile gignore.IgnoreFile
err := repo.Load(".gitignore", &ignoreFile)

// Find conflicts
conflicts := ignoreFile.FindConflicts()
for _, conflict := range conflicts {
    fmt.Printf("Conflict: %s between '%s' and '%s'\n",
        conflict.ConflictType,
        conflict.Left.Render(),
        conflict.Right.Render())
}
```

### Rule Reordering

```go
// Move a rule before or after another rule
result, err := service.MoveRule(".gitignore", "!important.txt", "build/**", gignore.AFTER)
```

### Parsing Existing Files

```go
// Parse .gitignore content
content := `*.log
build/
!build/important.txt`

var ignoreFile gignore.IgnoreFile
err := gignore.ParseIgnoreFile(content, &ignoreFile)
```

## Error Handling

The library uses explicit error types for better error handling:

```go
err := service.AddFileRule(".gitignore", "config.json", gignore.INCLUDE)
if err != nil {
    switch {
    case errors.Is(err, gignore.SemanticConflictError):
        // Handle semantic conflict
    case errors.Is(err, gignore.RedundantRuleError):
        // Handle redundant rule
    case errors.Is(err, gignore.UnreachableRuleError):
        // Handle unreachable rule
    }
}
```

## Testing

The library includes comprehensive test coverage and provides utilities for testing:

```go
// Create fake repository for testing
repo := gignore.NewFakeRepository()
service := gignore.NewService(repo)

// Test your ignore file logic
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

