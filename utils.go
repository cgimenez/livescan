package main

import (
	"fmt"
	"os"
)

// From https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func isIconFile(filename string) bool {
	var a = [...]byte{73, 99, 111, 110, 13}
	var b = []byte(filename)

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
