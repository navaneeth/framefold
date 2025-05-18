# Framefold

A minimal photo and video organizer written in Go, perfect for Raspberry Pi.

## Features

- Organizes photos and videos into customizable folder structures using Go templates
- Reads EXIF data for accurate date-based organization
- Falls back to file modification date when EXIF is not available
- Supports nested directory scanning
- Configurable media type handling
- Progress logging
- Custom file naming options
- Safe by default: copies files instead of moving them
- Processing summary with file counts, size, and timing in JSON format

## Installation

### From Source

1. Clone the repository
```bash
git clone https://github.com/yourusername/framefold.git
cd framefold
```

2. Install
```bash
make  # Compiles and installs to your Go bin directory
```

For local development, you can build without installing:
```bash
make local  # Creates framefold binary in current directory
```

### Release Builds

To build release tarballs for all supported platforms:
```bash
make release
```

This will create tarballs in the `dist` directory for:
- macOS (Intel and Apple Silicon)
- Linux (AMD64 and ARM64)
- Raspberry Pi (ARM v6 and v7)

Each tarball contains:
- The platform-specific binary
- README.md
- LICENSE (if present)
- Example configuration file (if present)

## Usage

Basic usage with default settings (copies files):
```bash
framefold --source ~/Pictures/Unsorted --target ~/Pictures/Organized
```

Copy files with custom configuration:
```bash
framefold --source ~/Pictures/Unsorted --target ~/Pictures/Organized --config ~/framefold-config.json
```

Move files instead of copying (deletes source files and empty directories):
```bash
framefold --source ~/Pictures/Unsorted --target ~/Pictures/Organized --delete-source
```

Show version information:
```bash
framefold --version
```

Example version output:
```
Framefold version 1.2.0 (a1b2c3d)  # If git tag v1.2.0 exists
# or
Framefold version dev (a1b2c3d)     # If no git tag exists
```

Example processing output:
```json
{
  "images": 120,
  "videos": 30,
  "files_with_exif": 115,
  "total_size_bytes": 2684354560,
  "total_files": 150,
  "duration": "2 minutes 15 seconds",
  "total_size": "2.5 GB"
}
```

## Configuration

Configuration is optional. If no config file is specified, the following default values will be used:

```json
{
    "folder_template": "{{.Year}}/{{.Month}}",
    "media_types": {
        "images": [".jpg", ".jpeg", ".png", ".gif", ".heic"],
        "videos": [".mp4", ".mov", ".avi"]
    },
    "use_original_filename": true,
    "logging": {
        "enabled": true,
        "level": "info"
    }
}
```

To customize the behavior, create a `config.json` file with your desired settings and pass it using the `--config` flag.

### Command Line Options

- `--source`: Source directory containing media files (required)
- `--target`: Target directory for organized files (required)
- `--config`: Path to configuration file (optional)
- `--delete-source`: Delete source files and empty directories after successful copy (optional, default: false)
- `--output`: Path to output file to write list of processed files (optional)
- `--version`: Show version information and git commit hash

### Template Variables

The following variables are available in the folder template:
- `{{.Year}}` - Four-digit year (e.g., "2025")
- `{{.Month}}` - Two-digit month (e.g., "04")
- `{{.Day}}` - Two-digit day (e.g., "24")
- `{{.Hour}}` - Two-digit hour in 24-hour format (e.g., "15")
- `{{.Minute}}` - Two-digit minute (e.g., "30")
- `{{.MediaType}}` - Type of media as defined in config (e.g., "images", "videos")
- `{{.Extension}}` - File extension without dot (e.g., "jpg", "mp4")

Example templates:
- `{{.Year}}/{{.Month}}` - Organizes by year/month (default)
- `{{.Year}}/{{.MediaType}}/{{.Month}}` - Organizes by year, then media type, then month
- `{{.MediaType}}/{{.Year}}/{{.Month}}-{{.Day}}` - Organizes by media type, then year, then month-day

## Operation

The program will:
1. Scan all files in the source directory (including subdirectories)
2. Read EXIF data from media files
3. Create folders in the target directory based on the template
4. Copy files to their corresponding folders
5. Delete source files and empty directories if --delete-source flag is used
6. Log the progress of each operation (if logging is enabled)
7. Display a summary of processed files in JSON format, including:
   - Total number of files processed
   - Number of images and videos
   - Number of files with EXIF data
   - Total size in bytes and human-readable format
   - Total time taken for processing
