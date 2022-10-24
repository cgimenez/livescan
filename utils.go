package main

import (
	"database/sql"
	"fmt"
	"os"

	imgui "github.com/AllenDang/giu/imgui"
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

func RGBA(r, g, b, a int) imgui.Vec4 {
	return imgui.Vec4{X: float32(r) / 255, Y: float32(g) / 255, Z: float32(b) / 255, W: float32(a) / 255}
}

func newDBCnx() (*sql.DB, error) {
	var cnx *sql.DB
	var db_path string
	var err error

	if appCtx.env == TEST {
		db_path = "testdata/db.sqlite"
	} else {
		db_path = "files/db.sqlite"
	}
	cnx, err = sql.Open("sqlite3", db_path)
	if err != nil {
		return nil, err
	}
	return cnx, nil
}

func UNUSED(x ...interface{}) {}
