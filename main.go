package main

import (
	"archive/zip"
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const helpText = `Archive Tool - Compress directories to ZIP format with confirmation

USAGE:
  archive [flags]

FLAGS:
  -h, --help              Show this help message and exit
  -o, --output string     Output zip filename
                          Default: archive_YYYYMMDD_HHMMSS.zip
  -e, --exclude string    Comma-separated patterns to exclude
                          Example: .git,.exe,node_modules
  -c, --compress string   Compression level: 'fast' (default) or 'best'
                          fast: Speed-optimized, larger file size
                          best: Size-optimized, slower compression
  -d, --delete            Delete original files after successful archiving
                          WARNING: This will permanently remove files!

EXAMPLES:
  Archive current directory (creates timestamped archive):
    archive

  Fast archiving (default):
    archive -o backup.zip

  Maximum compression (slower):
    archive --compress best -o backup.zip

  Create archive with custom filename:
    archive -o mybackup.zip
    archive --output mybackup.zip

  Exclude specific files and folders:
    archive -e ".git,.exe,node_modules"
    archive --exclude ".git,.exe,node_modules"

  Delete original files after archiving:
    archive -d
    archive --delete
    archive -o backup.zip -d
    archive --output backup.zip --exclude ".git" --delete

  Combine all options:
    archive -o backup.zip -e ".git,.DS_Store" -c best -d
    archive --output backup.zip --exclude ".git" --compress best --delete

NOTES:
  - You will be asked to confirm before the archive is created
  - Files already named archive*.zip will not be included in the archive
  - Directory structure is preserved in the ZIP file
  - With --delete flag, you will be asked for additional confirmation before deletion
  - Excluded files will NOT be deleted even with --delete flag
  - Compression is applied per-file, not to the entire archive

For more information, visit: https://github.com/archive-tool
`

func main() {
	outputFlag := flag.String("o", "", "Output zip filename (default: archive_TIMESTAMP.zip)")
	outputLongFlag := flag.String("output", "", "Output zip filename (default: archive_TIMESTAMP.zip)")
	excludeFlag := flag.String("e", "", "Comma-separated patterns to exclude (e.g., '.git,.exe,node_modules')")
	excludeLongFlag := flag.String("exclude", "", "Comma-separated patterns to exclude (e.g., '.git,.exe,node_modules')")
	compressFlag := flag.String("c", "fast", "Compression level: 'fast' (default) or 'best'")
	compressLongFlag := flag.String("compress", "fast", "Compression level: 'fast' (default) or 'best'")
	deleteFlag := flag.Bool("d", false, "Delete original files after successful archiving")
	deleteLongFlag := flag.Bool("delete", false, "Delete original files after successful archiving")
	helpFlag := flag.Bool("help", false, "Show help message")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, helpText)
	}

	flag.Parse()

	// Check for -h, --help, or help flag
	if *helpFlag {
		fmt.Print(helpText)
		os.Exit(0)
	}

	// Determine delete flag (prefer long flag if both are set)
	shouldDelete := *deleteFlag || *deleteLongFlag

	// Determine compression level (prefer long flag if both are set)
	compressLevel := *compressFlag
	if *compressLongFlag != "fast" {
		compressLevel = *compressLongFlag
	}

	// Validate compression level
	compressLevel = strings.ToLower(strings.TrimSpace(compressLevel))
	if compressLevel != "fast" && compressLevel != "best" {
		fmt.Fprintf(os.Stderr, "Error: Invalid compression level '%s'. Use 'fast' or 'best'\n", compressLevel)
		os.Exit(1)
	}

	// Get the current working directory
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get current directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Archive Tool\n")
	fmt.Printf("Current directory: %s\n\n", workDir)

	// Determine output filename (prefer long flag if both are set)
	outputFile := *outputFlag
	if *outputLongFlag != "" {
		outputFile = *outputLongFlag
	}
	if outputFile == "" {
		timestamp := time.Now().Format("20060102_150405")
		outputFile = fmt.Sprintf("archive_%s.zip", timestamp)
	}

	// Ensure .zip extension
	if !strings.HasSuffix(strings.ToLower(outputFile), ".zip") {
		outputFile += ".zip"
	}

	// Determine exclude patterns (prefer long flag if both are set)
	excludeValue := *excludeFlag
	if *excludeLongFlag != "" {
		excludeValue = *excludeLongFlag
	}

	// Parse exclusion patterns
	excludePatterns := []string{}
	if excludeValue != "" {
		excludePatterns = strings.Split(excludeValue, ",")
		for i := range excludePatterns {
			excludePatterns[i] = strings.TrimSpace(excludePatterns[i])
		}
	}

	// Check if directory is empty
	isEmpty, _ := isDirectoryEmpty(workDir, excludePatterns)
	if isEmpty {
		fmt.Println("❌ Directory is empty! No files to archive.")
		os.Exit(0)
	}

	// Display summary
	fmt.Printf("Archive Summary:\n")
	fmt.Printf("  Source directory: %s\n", workDir)
	fmt.Printf("  Output file: %s\n", outputFile)
	fmt.Printf("  Compression level: %s\n", compressLevel)
	if len(excludePatterns) > 0 {
		fmt.Printf("  Exclude patterns: %v\n", excludePatterns)
	}
	if shouldDelete {
		fmt.Printf("  ⚠️  Delete mode: ON (files will be deleted after archiving)\n")
	}
	fmt.Printf("\n")

	// Request confirmation
	if !confirm("Do you want to proceed? (yes/no): ") {
		fmt.Println("Archive cancelled.")
		os.Exit(0)
	}

	// Create archive
	archivedFiles := []string{}
	err = createArchive(workDir, outputFile, excludePatterns, &archivedFiles, compressLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create archive: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Archive created successfully: %s\n", outputFile)

	// Delete files and subdirectories if flag is set
	if shouldDelete && len(archivedFiles) > 0 {
		if !confirm("⚠️  WARNING: Delete original files permanently? (yes/no): ") {
			fmt.Println("Files not deleted. Archive kept.")
			os.Exit(0)
		}

		deletedCount := 0
		totalToDelete := len(archivedFiles)
		fmt.Println("\nDeleting files:")

		for i, filePath := range archivedFiles {
			err := os.Remove(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nWarning: Failed to delete %s: %v\n", filePath, err)
			} else {
				deletedCount++
			}
			printProgressBar(i+1, totalToDelete)
		}

		fmt.Printf("✓ Files deleted: %d\n", deletedCount)

		// Remove empty subdirectories (deepest first)
		dirsRemoved := removeEmptyDirs(workDir, outputFile, excludePatterns)
		if dirsRemoved > 0 {
			fmt.Printf("✓ Empty directories removed: %d\n", dirsRemoved)
		}

		fmt.Println("✓ Cleanup completed successfully!")
	}
}

var stdinReader = bufio.NewReader(os.Stdin)

func confirm(prompt string) bool {
	for {
		fmt.Print(prompt)
		response, err := stdinReader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			return false
		}

		response = strings.ToLower(strings.TrimSpace(response))
		switch response {
		case "yes", "y":
			return true
		case "no", "n":
			return false
		default:
			fmt.Println("Please enter 'yes' or 'no'.")
		}
	}
}

func printProgressBar(current, total int) {
	if total <= 0 {
		return
	}

	barWidth := 40
	percent := float64(current) / float64(total)
	filledWidth := int(percent * float64(barWidth))

	bar := "\r["
	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			bar += "="
		} else if i == filledWidth && current < total {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += fmt.Sprintf("] %d/%d (%.0f%%)", current, total, percent*100)

	fmt.Print(bar)
	if current == total {
		fmt.Println()
	}
}

func isDirectoryEmpty(sourceDir string, excludePatterns []string) (bool, int) {
	fileCount := 0
	filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Count only files, not directories
		if !info.IsDir() && !shouldExclude(path, sourceDir, excludePatterns) {
			fileCount++
		}
		return nil
	})
	return fileCount == 0, fileCount
}

func createArchive(sourceDir, outputFile string, excludePatterns []string, archivedFiles *[]string, compressLevel string) error {
	// First pass: count total files (must mirror the second pass logic)
	fmt.Println("Scanning directory...")
	totalFiles := 0
	filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == filepath.Join(sourceDir, outputFile) {
			return nil
		}
		if shouldExclude(path, sourceDir, excludePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			totalFiles++
		}
		return nil
	})

	fmt.Printf("Found %d files to archive\n\n", totalFiles)

	outputPath := filepath.Join(sourceDir, outputFile)
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Set compression method based on level
	var compressionMethod uint16
	if compressLevel == "best" {
		compressionMethod = zip.Deflate
	} else {
		// "fast" uses Store (no compression) which is fastest
		compressionMethod = zip.Store
	}

	fileCount := 0
	fmt.Println("Archiving files:")

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the output file itself
		if path == filepath.Join(sourceDir, outputFile) {
			return nil
		}

		// Check exclusion patterns
		if shouldExclude(path, sourceDir, excludePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Add directory entry
			relPath = filepath.ToSlash(relPath) + "/"
			_, err := zipWriter.Create(relPath)
			return err
		}

		// Add file entry with appropriate compression
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(relPath)
		header.Method = compressionMethod

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		if err == nil {
			fileCount++
			// Track archived file for potential deletion
			if archivedFiles != nil {
				*archivedFiles = append(*archivedFiles, path)
			}
			// Show progress
			printProgressBar(fileCount, totalFiles)
		}
		return err
	})

	if err != nil {
		return err
	}

	fmt.Printf("\n✓ Files archived: %d\n", fileCount)
	return nil
}

func shouldExclude(path, baseDir string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	relPath, err := filepath.Rel(baseDir, path)
	if err != nil {
		return false
	}

	for _, pattern := range patterns {
		// Check if pattern matches the filename or directory name
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
		// Check if pattern is contained in the path
		if strings.Contains(relPath, pattern) {
			return true
		}
	}

	return false
}

// removeEmptyDirs walks the directory tree bottom-up and removes empty
// subdirectories left behind after file deletion. It never removes the
// root directory itself or excluded directories.
func removeEmptyDirs(rootDir, outputFile string, excludePatterns []string) int {
	removed := 0

	// Collect all subdirectories (deepest first via reverse order)
	var dirs []string
	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() || path == rootDir {
			return nil
		}
		if shouldExclude(path, rootDir, excludePatterns) {
			return filepath.SkipDir
		}
		dirs = append(dirs, path)
		return nil
	})

	// Reverse so deepest directories are removed first
	for i, j := 0, len(dirs)-1; i < j; i, j = i+1, j-1 {
		dirs[i], dirs[j] = dirs[j], dirs[i]
	}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		if len(entries) == 0 {
			if os.Remove(dir) == nil {
				removed++
			}
		}
	}

	return removed
}
