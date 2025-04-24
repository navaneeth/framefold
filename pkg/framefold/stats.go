package framefold

import (
	"fmt"
	"time"
)

// Stats tracks processing statistics
type Stats struct {
	ImageCount     int64
	VideoCount     int64
	ExifFound      int64
	TotalSize      int64
	ProcessedFiles int64
	StartTime      time.Time
}

// String formats the stats for display
func (s Stats) String() string {
	duration := time.Since(s.StartTime)
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60

	return fmt.Sprintf(`
Processing Summary:
-----------------
Total Files Processed: %d
Images: %d
Videos: %d
Files with EXIF data: %d
Total Size: %s
Time Taken: %d minutes %d seconds
`, s.ProcessedFiles, s.ImageCount, s.VideoCount, s.ExifFound, formatSize(s.TotalSize), minutes, seconds)
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
