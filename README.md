# Archive Tool

A powerful command-line utility written in Go to archive all files in the current directory to a `.zip` file with CLI confirmation, progress tracking, and optional cleanup.

## Features

- ✅ Recursively archives entire directory structure to ZIP format
- ✅ Interactive CLI confirmation before archiving
- ✅ **Real-time progress bars** for archiving and deletion
- ✅ Auto-generates timestamped archive names
- ✅ Customizable output filename (`-o`, `--output`)
- ✅ Support for excluding file/folder patterns (`-e`, `--exclude`)
- ✅ **Dual compression modes**: Fast (default) or Best compression
- ✅ **Optional file deletion after archiving** to free up space (`-d`, `--delete`)
- ✅ **Automatic empty directory cleanup** after deletion
- ✅ **Empty directory detection** - prevents archiving empty folders
- ✅ Comprehensive help documentation (`-h`, `--help`)
- ✅ Cross-platform (Windows, macOS, Linux)
- ✅ No external dependencies - pure Go standard library

## Installation

### Prerequisites
- Go 1.26 or later

### Build

```powershell
# Navigate to the project directory
cd d:\GoGoGo\archive-tool

# Build the executable
go build -o archive.exe
```

### Add to PATH (Optional)

1. Copy `archive.exe` to a directory in your PATH (e.g., `C:\Program Files\ArchiveTool\`)
2. Or add the project directory to your PATH

## Usage

### Basic Usage

```cmd
archive.exe
```

This will create an archive named `archive_YYYYMMDD_HHMMSS.zip` in the current directory with all files and subdirectories.

### Custom Output Filename

```cmd
archive.exe -o myarchive.zip
archive.exe --output myarchive.zip
```

### Exclude Patterns

```cmd
archive.exe -e ".git,.exe,node_modules"
archive.exe --exclude ".git,.exe,node_modules"
```

This will exclude files/directories matching `.git`, `.exe`, or `node_modules`.

### Delete Files After Archiving

```cmd
archive.exe -d
archive.exe --delete
```

This will delete original files and clean up empty subdirectories after successful archiving to free up space. You'll be asked for confirmation before deletion.

### Compression Modes

```cmd
# Fast compression (default) - Speed-optimized, faster but larger files
archive.exe -o backup.zip
archive.exe --compress fast

# Best compression - Size-optimized, slower but smaller archive files
archive.exe --compress best -o backup.zip
archive.exe -c best -o myfiles.zip
```

**Compression Mode Comparison:**
- **fast (default)**: Uses no compression (Store method), fastest archiving speed
- **best**: Uses DEFLATE compression, produces smaller archives but slower

### Combined Examples

```cmd
# Archive with custom name and exclude patterns
archive.exe -o backup.zip -e ".git,.DS_Store,__pycache__"

# Archive with custom name and delete original files
archive.exe --output backup.zip -d

# Archive with best compression (smaller file, slower)
archive.exe -c best -o backup.zip

# Archive with all options combined
archive.exe --output final_backup.zip --exclude ".git,.env" --compress best --delete
```

### Show Help

```cmd
archive.exe -h
archive.exe --help
```

## How It Works

1. **Validates Directory** - Checks if the directory has files to archive
2. **Displays Summary** - Shows the current directory, output filename, compression level, and options
3. **Requests Confirmation** - Asks user to confirm before proceeding
4. **Scans Directory** - Recursively counts all files to display accurate progress
5. **Creates Archive** - Archives all files and subdirectories with real-time progress bar
6. **Optional Cleanup** - If delete flag is set, asks for confirmation then deletes original files and cleans up empty subdirectories
7. **Reports Results** - Shows final statistics and success message

## Command-Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-h`, `--help` | Show help message and exit | N/A |
| `-o`, `--output` | Output zip filename | `archive_YYYYMMDD_HHMMSS.zip` |
| `-e`, `--exclude` | Comma-separated exclusion patterns | (none) |
| `-c`, `--compress` | Compression level: `fast` or `best` | `fast` |
| `-d`, `--delete` | Delete original files and empty dirs | (disabled) |

## Progress Bars

The tool displays real-time progress bars for both archiving and deletion:

```
Archiving files:
[=============================>          ] 10/15 (67%)

Deleting files:
[=============================>          ] 10/15 (67%)
```

Progress bars show:
- Visual progress indicator
- Current file count and total files
- Percentage completion

## Compression Modes

### Fast Mode (Default)
- **Method**: Store (no compression)
- **Speed**: Very fast
- **File size**: Larger (original sizes)
- **Use case**: Quick backups, already compressed files (images, videos, zips)
- **Example**: `archive.exe -o backup.zip` or `archive.exe -c fast`

### Best Mode
- **Method**: DEFLATE compression algorithm
- **Speed**: Slower than fast mode
- **File size**: Smaller (can reduce by 50-90% for text files)
- **Use case**: Maximum space savings, archiving source code, documents
- **Example**: `archive.exe --compress best -o backup.zip`

**Performance Comparison:**
- Fast mode: Ideal for speed - archives complete quickly
- Best mode: Ideal for size - produces compact archives but takes longer

Choose based on your priorities:
- **Need speed?** Use `fast` (default)
- **Need smaller files?** Use `best`

## Example Workflow

### Fast Compression (Default)

```cmd
C:\MyProject> archive.exe -o project_backup.zip
Archive Tool
Current directory: C:\MyProject

Archive Summary:
  Source directory: C:\MyProject
  Output file: project_backup.zip
  Compression level: fast

Do you want to proceed? (yes/no): yes

Scanning directory...
Found 42 files to archive

Archiving files:
[========================================] 42/42 (100%)

✓ Files archived: 42
✓ Archive created successfully: project_backup.zip
```

### Best Compression (Smaller File Size)

```cmd
C:\MyProject> archive.exe -o project_backup.zip --compress best
Archive Tool
Current directory: C:\MyProject

Archive Summary:
  Source directory: C:\MyProject
  Output file: project_backup.zip
  Compression level: best

Do you want to proceed? (yes/no): yes

Scanning directory...
Found 42 files to archive

Archiving files:
[========================================] 42/42 (100%)

✓ Files archived: 42
✓ Archive created successfully: project_backup.zip
```

### With Delete (Full Cleanup)

```cmd
C:\MyProject> archive.exe -o backup.zip -d
Archive Tool
Current directory: C:\MyProject

Archive Summary:
  Source directory: C:\MyProject
  Output file: backup.zip
  Compression level: fast
  ⚠️  Delete mode: ON (files will be deleted after archiving)

Do you want to proceed? (yes/no): yes

Scanning directory...
Found 35 files to archive

Archiving files:
[========================================] 35/35 (100%)

✓ Files archived: 35
✓ Archive created successfully: backup.zip

⚠️  WARNING: Delete original files permanently? (yes/no): yes

Deleting files:
[========================================] 35/35 (100%)
✓ Files deleted: 35
✓ Empty directories removed: 8
✓ Cleanup completed successfully!
```

### Empty Directory

```cmd
C:\EmptyFolder> archive.exe
Archive Tool
Current directory: C:\EmptyFolder

❌ Directory is empty! No files to archive.
```

## Safety Features

- ✅ Two-level confirmation (archive + delete separately)
- ✅ Excluded files and directories are NOT deleted even with `--delete` flag
- ✅ Archive file itself is NOT included in the archive
- ✅ Empty directory detection prevents unnecessary operations
- ✅ Automatic empty directory cleanup after file deletion
- ✅ Warning messages for destructive operations
- ✅ Graceful error handling with informative messages

## Notes

- Recursively archives all files and subdirectories (uses `filepath.Walk`)
- Complete directory structure is fully preserved in the ZIP file
- With `--delete`, files are deleted first, then empty subdirectories are cleaned up (bottom-up)
- Excluded files and directories will NOT be deleted even with `--delete` flag
- The tool respects the same exclusion patterns during both archiving and deletion
- Only operates on the current working directory (not recursive from parent directories)

## License

MIT
