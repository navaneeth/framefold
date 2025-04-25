package framefold

import (
	"encoding/json"
	"fmt"
	"time"
)

// Stats tracks processing statistics
type Stats struct {
	ImageCount     int64     `json:"images"`
	VideoCount     int64     `json:"videos"`
	ExifFound      int64     `json:"files_with_exif"`
	TotalSize      int64     `json:"total_size_bytes"`
	ProcessedFiles int64     `json:"total_files"`
	StartTime      time.Time `json:"-"` // Don't include in JSON output
	Duration      string    `json:"duration"`
	HumanSize     string    `json:"total_size"`
}

// String formats the stats as JSON
func (s Stats) String() string {
	// Calculate duration
	duration := time.Since(s.StartTime)
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60
	s.Duration = fmt.Sprintf("%d minutes %d seconds", minutes, seconds)
	
	// Format human-readable size
	s.HumanSize = formatSize(s.TotalSize)

	// Marshal to JSON
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Sprintf("error formatting stats: %v", err)
	}

	return string(data)
}

// formatSize converts bytes to human-readable format
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
