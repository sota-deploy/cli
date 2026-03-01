package archive

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// defaultIgnores are always excluded from archives.
var defaultIgnores = []string{
	".git/",
	".git",
	"node_modules/",
	"node_modules",
	".env",
	".env.*",
	"*.log",
	".DS_Store",
	".sota.json",
	"__pycache__/",
	"__pycache__",
	".venv/",
	".venv",
	"vendor/",
}

// Create creates a tar.gz archive of the given directory, respecting ignore patterns.
// Returns the archive as a temporary file (caller must close and remove).
func Create(dir string) (*os.File, int64, error) {
	// Load ignore patterns
	patterns := loadIgnorePatterns(dir)

	// Create temp file for the archive
	tmpFile, err := os.CreateTemp("", "sota-archive-*.tar.gz")
	if err != nil {
		return nil, 0, fmt.Errorf("creating temp file: %w", err)
	}

	gzWriter := gzip.NewWriter(tmpFile)
	tarWriter := tar.NewWriter(gzWriter)

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		// Check if path should be ignored
		if shouldIgnore(relPath, info.IsDir(), patterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("creating header for %s: %w", relPath, err)
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("writing header for %s: %w", relPath, err)
		}

		// Write file content (skip directories)
		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("opening %s: %w", relPath, err)
			}
			defer f.Close()

			if _, err := io.Copy(tarWriter, f); err != nil {
				return fmt.Errorf("copying %s: %w", relPath, err)
			}
		}

		return nil
	})

	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, 0, fmt.Errorf("creating archive: %w", err)
	}

	if err := tarWriter.Close(); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, 0, err
	}
	if err := gzWriter.Close(); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, 0, err
	}

	// Get file size
	info, err := tmpFile.Stat()
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, 0, err
	}

	// Seek to beginning for reading
	if _, err := tmpFile.Seek(0, 0); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, 0, err
	}

	return tmpFile, info.Size(), nil
}

// loadIgnorePatterns loads patterns from .sotaignore, falling back to .gitignore.
func loadIgnorePatterns(dir string) []string {
	patterns := make([]string, len(defaultIgnores))
	copy(patterns, defaultIgnores)

	// Try .sotaignore first, then .gitignore
	for _, name := range []string{".sotaignore", ".gitignore"} {
		path := filepath.Join(dir, name)
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			patterns = append(patterns, line)
		}
		break // Use first file found
	}

	return patterns
}

// shouldIgnore checks if a path matches any ignore pattern.
func shouldIgnore(relPath string, isDir bool, patterns []string) bool {
	for _, pattern := range patterns {
		// Directory patterns (ending with /)
		if strings.HasSuffix(pattern, "/") {
			dirPattern := strings.TrimSuffix(pattern, "/")
			if isDir && (relPath == dirPattern || strings.HasPrefix(relPath, dirPattern+"/")) {
				return true
			}
			if !isDir && strings.HasPrefix(relPath, dirPattern+"/") {
				return true
			}
			continue
		}

		// Exact match
		if relPath == pattern {
			return true
		}

		// Base name match
		base := filepath.Base(relPath)
		if base == pattern {
			return true
		}

		// Simple wildcard matching
		if strings.Contains(pattern, "*") {
			if matched, _ := filepath.Match(pattern, base); matched {
				return true
			}
			if matched, _ := filepath.Match(pattern, relPath); matched {
				return true
			}
		}

		// Prefix match for directory components
		parts := strings.Split(relPath, string(filepath.Separator))
		for _, part := range parts {
			if part == pattern {
				return true
			}
		}
	}
	return false
}

// FormatSize formats a byte count as a human-readable string.
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
	)
	switch {
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
