package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create nested structure:
	//   root.txt
	//   sub1/a.txt
	//   sub1/deep/b.txt
	//   sub2/c.txt
	//   excluded/.secret
	os.MkdirAll(filepath.Join(dir, "sub1", "deep"), 0755)
	os.MkdirAll(filepath.Join(dir, "sub2"), 0755)
	os.MkdirAll(filepath.Join(dir, "excluded"), 0755)

	files := map[string]string{
		"root.txt":            "root content",
		"sub1/a.txt":          "sub1 content",
		"sub1/deep/b.txt":     "deep content",
		"sub2/c.txt":          "sub2 content",
		"excluded/.secret":    "secret",
	}

	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", name, err)
		}
	}

	return dir
}

func TestShouldExclude(t *testing.T) {
	baseDir := "/project"

	tests := []struct {
		name     string
		path     string
		patterns []string
		want     bool
	}{
		{"no patterns", "/project/file.txt", nil, false},
		{"substring match", "/project/file.exe", []string{".exe"}, true},
		{"match extension glob", "/project/file.exe", []string{"*.exe"}, true},
		{"match directory name", "/project/.git/config", []string{".git"}, true},
		{"no match", "/project/src/main.go", []string{".git", "node_modules"}, false},
		{"match nested", "/project/node_modules/pkg/index.js", []string{"node_modules"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldExclude(tt.path, baseDir, tt.patterns)
			if got != tt.want {
				t.Errorf("shouldExclude(%q, %q, %v) = %v, want %v", tt.path, baseDir, tt.patterns, got, tt.want)
			}
		})
	}
}

func TestIsDirectoryEmpty(t *testing.T) {
	t.Run("empty directory", func(t *testing.T) {
		dir := t.TempDir()
		isEmpty, count := isDirectoryEmpty(dir, nil)
		if !isEmpty {
			t.Errorf("expected empty directory, got count=%d", count)
		}
	})

	t.Run("directory with files", func(t *testing.T) {
		dir := setupTestDir(t)
		isEmpty, count := isDirectoryEmpty(dir, nil)
		if isEmpty {
			t.Errorf("expected non-empty directory")
		}
		if count != 5 {
			t.Errorf("expected 5 files, got %d", count)
		}
	})

	t.Run("directory with exclusions", func(t *testing.T) {
		dir := setupTestDir(t)
		isEmpty, count := isDirectoryEmpty(dir, []string{"excluded"})
		if isEmpty {
			t.Errorf("expected non-empty directory")
		}
		if count != 4 {
			t.Errorf("expected 4 files after exclusion, got %d", count)
		}
	})

	t.Run("all files excluded", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".git"), 0755)
		os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("data"), 0644)

		isEmpty, _ := isDirectoryEmpty(dir, []string{".git"})
		if !isEmpty {
			t.Errorf("expected empty after excluding all files")
		}
	})
}

func TestCreateArchive(t *testing.T) {
	t.Run("fast compression archives all files", func(t *testing.T) {
		dir := setupTestDir(t)
		outputFile := "test_archive.zip"
		outputPath := filepath.Join(dir, outputFile)

		var archivedFiles []string
		err := createArchive(dir, outputFile, nil, &archivedFiles, "fast")
		if err != nil {
			t.Fatalf("createArchive failed: %v", err)
		}

		if len(archivedFiles) != 5 {
			t.Errorf("expected 5 archived files, got %d", len(archivedFiles))
		}

		verifyZipContents(t, outputPath, 5)
	})

	t.Run("best compression archives all files", func(t *testing.T) {
		dir := setupTestDir(t)
		outputFile := "test_archive.zip"
		outputPath := filepath.Join(dir, outputFile)

		var archivedFiles []string
		err := createArchive(dir, outputFile, nil, &archivedFiles, "best")
		if err != nil {
			t.Fatalf("createArchive failed: %v", err)
		}

		if len(archivedFiles) != 5 {
			t.Errorf("expected 5 archived files, got %d", len(archivedFiles))
		}

		verifyZipContents(t, outputPath, 5)

		// Verify deflate compression is used
		reader, err := zip.OpenReader(outputPath)
		if err != nil {
			t.Fatalf("failed to open zip: %v", err)
		}
		defer reader.Close()

		for _, f := range reader.File {
			if !f.FileInfo().IsDir() && f.Method != zip.Deflate {
				t.Errorf("expected Deflate for %s, got method %d", f.Name, f.Method)
			}
		}
	})

	t.Run("excludes patterns", func(t *testing.T) {
		dir := setupTestDir(t)
		outputFile := "test_archive.zip"
		outputPath := filepath.Join(dir, outputFile)

		var archivedFiles []string
		err := createArchive(dir, outputFile, []string{"excluded"}, &archivedFiles, "fast")
		if err != nil {
			t.Fatalf("createArchive failed: %v", err)
		}

		if len(archivedFiles) != 4 {
			t.Errorf("expected 4 archived files (1 excluded), got %d", len(archivedFiles))
		}

		verifyZipContents(t, outputPath, 4)
	})

	t.Run("preserves directory structure", func(t *testing.T) {
		dir := setupTestDir(t)
		outputFile := "test_archive.zip"
		outputPath := filepath.Join(dir, outputFile)

		err := createArchive(dir, outputFile, nil, nil, "fast")
		if err != nil {
			t.Fatalf("createArchive failed: %v", err)
		}

		reader, err := zip.OpenReader(outputPath)
		if err != nil {
			t.Fatalf("failed to open zip: %v", err)
		}
		defer reader.Close()

		expectedFiles := map[string]bool{
			"root.txt":         false,
			"sub1/a.txt":       false,
			"sub1/deep/b.txt":  false,
			"sub2/c.txt":       false,
			"excluded/.secret": false,
		}

		for _, f := range reader.File {
			if f.FileInfo().IsDir() {
				continue
			}
			if _, ok := expectedFiles[f.Name]; ok {
				expectedFiles[f.Name] = true
			}
		}

		for name, found := range expectedFiles {
			if !found {
				t.Errorf("expected file %q not found in archive", name)
			}
		}
	})

	t.Run("skips output file", func(t *testing.T) {
		dir := setupTestDir(t)
		outputFile := "test_archive.zip"
		outputPath := filepath.Join(dir, outputFile)

		err := createArchive(dir, outputFile, nil, nil, "fast")
		if err != nil {
			t.Fatalf("createArchive failed: %v", err)
		}

		reader, err := zip.OpenReader(outputPath)
		if err != nil {
			t.Fatalf("failed to open zip: %v", err)
		}
		defer reader.Close()

		for _, f := range reader.File {
			if f.Name == outputFile {
				t.Errorf("output file %q should not be in the archive", outputFile)
			}
		}
	})
}

func TestRemoveEmptyDirs(t *testing.T) {
	t.Run("removes empty subdirectories", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "a", "b", "c"), 0755)
		os.MkdirAll(filepath.Join(dir, "d"), 0755)

		removed := removeEmptyDirs(dir, "", nil)
		if removed != 4 {
			t.Errorf("expected 4 empty dirs removed, got %d", removed)
		}

		entries, _ := os.ReadDir(dir)
		if len(entries) != 0 {
			t.Errorf("expected empty root dir, got %d entries", len(entries))
		}
	})

	t.Run("keeps non-empty directories", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "keep"), 0755)
		os.WriteFile(filepath.Join(dir, "keep", "file.txt"), []byte("data"), 0644)
		os.MkdirAll(filepath.Join(dir, "remove"), 0755)

		removed := removeEmptyDirs(dir, "", nil)
		if removed != 1 {
			t.Errorf("expected 1 empty dir removed, got %d", removed)
		}

		if _, err := os.Stat(filepath.Join(dir, "keep")); os.IsNotExist(err) {
			t.Error("'keep' directory should not have been removed")
		}
	})

	t.Run("respects exclusion patterns", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".git"), 0755)
		os.MkdirAll(filepath.Join(dir, "empty"), 0755)

		removed := removeEmptyDirs(dir, "", []string{".git"})
		if removed != 1 {
			t.Errorf("expected 1 empty dir removed (skipping .git), got %d", removed)
		}

		if _, err := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
			t.Error("excluded .git directory should not have been removed")
		}
	})
}

func verifyZipContents(t *testing.T, zipPath string, expectedFileCount int) {
	t.Helper()

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open zip %s: %v", zipPath, err)
	}
	defer reader.Close()

	fileCount := 0
	for _, f := range reader.File {
		if !f.FileInfo().IsDir() {
			fileCount++
		}
	}

	if fileCount != expectedFileCount {
		t.Errorf("expected %d files in zip, got %d", expectedFileCount, fileCount)
	}
}
