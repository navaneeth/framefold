package framefold

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"
)

const (
	// Buffer size for file operations (1MB)
	copyBufferSize = 1024 * 1024
)

// FileInfo holds template variables for folder organization
type FileInfo struct {
	Year      string
	Month     string
	Day       string
	Hour      string
	Minute    string
	MediaType string
	Extension string
}

// Processor handles the file organization process
type Processor struct {
	config            Config
	stats             Stats
	sourceDir         string
	targetDir         string
	deleteSource      bool
	processedDirs     map[string]bool // Track directories that had files processed
	processedFiles    []string        // Track processed files for output
	outputPath        string          // Path to output file
	lock              *processLock
	exiftoolChecked   bool // Track if exiftool availability has been checked
	exiftoolAvailable bool // Whether exiftool is available
}

// NewProcessor creates a new file processor
func NewProcessor(sourceDir, targetDir string, config Config, deleteSource bool, outputPath string) (*Processor, error) {
	lock, err := newProcessLock()
	if err != nil {
		return nil, fmt.Errorf("failed to create process lock: %v", err)
	}

	return &Processor{
		config:         config,
		sourceDir:      sourceDir,
		targetDir:      targetDir,
		deleteSource:   deleteSource,
		stats:          Stats{StartTime: time.Now()},
		processedDirs:  make(map[string]bool),
		processedFiles: make([]string, 0),
		outputPath:     outputPath,
		lock:           lock,
	}, nil
}

// Process organizes files from source to target directory
// WriteProcessedFiles writes the list of processed files to the output file
func (p *Processor) WriteProcessedFiles() error {
	if p.outputPath == "" {
		return nil
	}

	// Sort files for consistent output
	sort.Strings(p.processedFiles)

	// Create output file
	f, err := os.Create(p.outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer f.Close()

	// Write each file path on a new line
	for _, file := range p.processedFiles {
		if _, err := fmt.Fprintln(f, file); err != nil {
			return fmt.Errorf("failed to write to output file: %v", err)
		}
	}

	return nil
}

func (p *Processor) Process() error {
	// Try to acquire lock
	locked, err := p.lock.acquire()
	if err != nil {
		return fmt.Errorf("error acquiring lock: %v", err)
	}
	if !locked {
		return fmt.Errorf("another instance of framefold is already running")
	}
	defer p.lock.release()

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(p.targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	// Walk through the source directory
	err = filepath.Walk(p.sourceDir, p.processFile)
	if err != nil {
		return fmt.Errorf("error walking through directory: %v", err)
	}

	// If deleting source files, clean up empty directories
	if p.deleteSource {
		if err := p.cleanEmptyDirs(); err != nil {
			return fmt.Errorf("error cleaning empty directories: %v", err)
		}
	}

	// Write processed files list if output path is set
	if err := p.WriteProcessedFiles(); err != nil {
		return fmt.Errorf("error writing processed files list: %v", err)
	}

	return nil
}

// cleanEmptyDirs removes empty directories in the source tree
func (p *Processor) cleanEmptyDirs() error {
	var dirsToCheck []string

	// Get all directories that had files processed
	for dir := range p.processedDirs {
		dirsToCheck = append(dirsToCheck, dir)
	}

	// Sort directories by depth (deepest first) to ensure proper removal
	// This ensures we process child directories before their parents
	for i := 0; i < len(dirsToCheck)-1; i++ {
		for j := i + 1; j < len(dirsToCheck); j++ {
			if len(strings.Split(dirsToCheck[i], string(os.PathSeparator))) < len(strings.Split(dirsToCheck[j], string(os.PathSeparator))) {
				dirsToCheck[i], dirsToCheck[j] = dirsToCheck[j], dirsToCheck[i]
			}
		}
	}

	// Check and remove empty directories
	for _, dir := range dirsToCheck {
		// Read directory contents
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Directory was already removed
			}
			return fmt.Errorf("error reading directory %s: %v", dir, err)
		}

		// If directory is empty, remove it
		if len(entries) == 0 {
			if err := os.Remove(dir); err != nil {
				return fmt.Errorf("error removing empty directory %s: %v", dir, err)
			}
			if p.config.Logging.Enabled {
				log.Printf("Removed empty directory: %s", dir)
			}
		}
	}

	return nil
}

// GetStats returns the current processing statistics
func (p *Processor) GetStats() Stats {
	return p.stats
}

func (p *Processor) processFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	// Skip directories
	if info.IsDir() {
		return nil
	}

	// Check if file is a supported media type
	ext := strings.ToLower(filepath.Ext(path))
	mediaType := p.getMediaType(ext)
	if mediaType == "" {
		return nil
	}

	// Track the directory containing this file
	p.processedDirs[filepath.Dir(path)] = true

	// Update media type counts
	p.stats.ProcessedFiles++
	if mediaType == "images" {
		p.stats.ImageCount++
	} else if mediaType == "videos" {
		p.stats.VideoCount++
	}

	// Update total size
	p.stats.TotalSize += info.Size()

	// Get file date
	date, err := p.getFileDate(path)
	if err != nil {
		if p.config.Logging.Enabled {
			log.Printf("Warning: Could not get EXIF data for %s, using file modification time", path)
		}
		date = info.ModTime()
	} else {
		p.stats.ExifFound++
	}

	// Create file info for template
	fileInfo := FileInfo{
		Year:      date.Format("2006"),
		Month:     date.Format("01"),
		Day:       date.Format("02"),
		Hour:      date.Format("15"),
		Minute:    date.Format("04"),
		MediaType: mediaType,
		Extension: ext[1:], // Remove the dot
	}

	// Parse and execute the template
	tmpl, err := template.New("folder").Parse(p.config.FolderTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	var targetPath strings.Builder
	if err := tmpl.Execute(&targetPath, fileInfo); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// Create the target directory
	newDir := filepath.Join(p.targetDir, targetPath.String())
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", newDir, err)
	}

	// Generate target filename
	var filename string
	if p.config.UseOriginalName {
		filename = filepath.Base(path)
	} else {
		filename = fmt.Sprintf("%s-%s-%s%s",
			date.Format("20060102"),
			date.Format("150405"),
			mediaType,
			ext)
	}

	// Copy file to new location
	newPath := filepath.Join(newDir, filename)

	// Check if target file exists and is identical
	if identical, err := p.areFilesIdentical(path, newPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error comparing files: %v", err)
		}
	} else if identical {
		if p.config.Logging.Enabled {
			log.Printf("Skipping identical file: %s", path)
		}
		if p.deleteSource {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to delete source file %s: %v", path, err)
			}
		}
		return nil
	}

	if err := p.copyFile(path, newPath); err != nil {
		return fmt.Errorf("failed to copy file %s: %v", path, err)
	}

	// Delete source file if requested
	if p.deleteSource {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to delete source file %s: %v", path, err)
		}
		if p.config.Logging.Enabled {
			log.Printf("Moved %s to %s", path, newPath)
		}
		// Add to processed files list
		p.processedFiles = append(p.processedFiles, newPath)
	} else {
		if p.config.Logging.Enabled {
			log.Printf("Copied %s to %s", path, newPath)
		}
		// Add to processed files list
		p.processedFiles = append(p.processedFiles, newPath)
	}

	return nil
}

// areFilesIdentical efficiently compares two files by size and hash
func (p *Processor) areFilesIdentical(src, dst string) (bool, error) {
	// First check if destination exists
	dstInfo, err := os.Stat(dst)
	if err != nil {
		return false, err
	}

	// Check if source exists
	srcInfo, err := os.Stat(src)
	if err != nil {
		return false, err
	}

	// Quick size comparison
	if srcInfo.Size() != dstInfo.Size() {
		return false, nil
	}

	// Compare file contents using SHA-256 hash
	srcHash, err := p.calculateFileHash(src)
	if err != nil {
		return false, err
	}

	dstHash, err := p.calculateFileHash(dst)
	if err != nil {
		return false, err
	}

	return srcHash == dstHash, nil
}

// calculateFileHash calculates SHA-256 hash of a file using buffered reads
func (p *Processor) calculateFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	buf := make([]byte, copyBufferSize)

	for {
		n, err := file.Read(buf)
		if n > 0 {
			hash.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (p *Processor) getMediaType(ext string) string {
	for mediaType, extensions := range p.config.MediaTypes {
		for _, e := range extensions {
			if e == ext {
				return mediaType
			}
		}
	}
	return ""
}

// checkExiftool verifies that exiftool is available on the system
func (p *Processor) checkExiftool() error {
	cmd := exec.Command("exiftool", "-ver")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("exiftool is not available: %v\nPlease install exiftool to extract EXIF data from media files", err)
	}
	return nil
}

func (p *Processor) getFileDate(path string) (time.Time, error) {
	// Check exiftool availability on first call
	if !p.exiftoolChecked {
		p.exiftoolChecked = true
		if err := p.checkExiftool(); err != nil {
			return time.Time{}, err
		}
		p.exiftoolAvailable = true
	}

	// Try to get DateTimeOriginal first, then DateTime as fallback
	cmd := exec.Command("exiftool", "-DateTimeOriginal", "-DateTime", "-s", "-s", "-s", path)
	output, err := cmd.Output()
	if err != nil {
		return time.Time{}, fmt.Errorf("exiftool execution failed: %v", err)
	}

	// Parse the output - exiftool returns the first matching tag
	dateTimeStr := strings.TrimSpace(string(output))
	if dateTimeStr == "" {
		return time.Time{}, fmt.Errorf("no DateTime found in EXIF")
	}

	// Parse the EXIF date format: "YYYY:MM:DD HH:MM:SS"
	return time.Parse("2006:01:02 15:04:05", dateTimeStr)
}

func (p *Processor) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	buf := make([]byte, copyBufferSize)
	_, err = io.CopyBuffer(dstFile, srcFile, buf)
	if err != nil {
		return err
	}

	// Copy file mode from source
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}
