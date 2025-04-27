package framefold

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const lockFileName = ".framefold.lock"

type processLock struct {
	path string
}

// newProcessLock creates a new process lock
func newProcessLock() (*processLock, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not get home directory: %v", err)
	}
	return &processLock{
		path: filepath.Join(homeDir, lockFileName),
	}, nil
}

// acquire tries to create a lock file
// Returns true if lock was acquired, false if already locked
func (l *processLock) acquire() (bool, error) {
	// Try to create lock file
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		if os.IsExist(err) {
			// Check if the lock file is stale
			if l.isStale() {
				// Remove stale lock and try again
				os.Remove(l.path)
				return l.acquire()
			}
			return false, nil // Lock exists and is not stale
		}
		return false, fmt.Errorf("error creating lock file: %v", err)
	}
	defer file.Close()

	// Write current PID to lock file
	pid := os.Getpid()
	if _, err := fmt.Fprintf(file, "%d", pid); err != nil {
		os.Remove(l.path) // Clean up on write error
		return false, fmt.Errorf("error writing to lock file: %v", err)
	}

	return true, nil
}

// release removes the lock file
func (l *processLock) release() error {
	// Only remove if it exists
	if err := os.Remove(l.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error removing lock file: %v", err)
	}
	return nil
}

// isStale checks if the lock file is stale
// A lock is considered stale if:
// 1. The file is older than 24 hours
// 2. The process ID in the file doesn't exist
func (l *processLock) isStale() bool {
	info, err := os.Stat(l.path)
	if err != nil {
		return false
	}

	// Check if lock is older than 24 hours
	if info.ModTime().Add(24 * time.Hour).Before(time.Now()) {
		return true
	}

	// Read PID from lock file
	data, err := os.ReadFile(l.path)
	if err != nil {
		return false
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return true // Invalid PID format, consider lock stale
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return true // Can't find process, consider lock stale
	}

	// On Unix systems, signal 0 can be used to test if process exists
	err = process.Signal(syscall.Signal(0))
	return err != nil // If error, process doesn't exist
}
