package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Ableton Packs dir already scanned ?
func isLivePacksDBScanned() (bool, error) {
	var count int
	stmt := "SELECT COUNT(*) FROM sqlite_master WHERE name ='packs_samples' and type='table'"
	_ = appCtx.dbCnx.QueryRow(stmt).Scan(&count)
	if count == 0 {
		stmt = "CREATE TABLE packs_samples (directory TEXT, name TEXT)"
		if _, err := appCtx.dbCnx.Exec(stmt); err != nil {
			return false, err
		}
		return false, nil
	} else {
		stmt = "SELECT COUNT(*) FROM packs_samples"
		_ = appCtx.dbCnx.QueryRow(stmt).Scan(&count)
		if count == 0 {
			return false, nil
		}
	}
	return true, nil
}

func scanLivePacks() error {
	var err error

	if appCtx.libRootPath == "" {
		return errors.New("libRootPath is not set")
	}

	if _, err = appCtx.dbCnx.Exec("DELETE FROM packs_samples"); err != nil {
		return err
	}

	err = filepath.Walk(appCtx.libRootPath,
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				fmt.Println("Scanning", path)
			}
			dir := filepath.Dir(path)
			base := filepath.Base(path)
			if err != nil && !isIconFile(base) { // Icon? files cause problems under Windows
				return fmt.Errorf("name %s: %v", base, err)
			}
			switch filepath.Ext(path) {
			case ".aif", ".wav", ".mp3":
				if _, err = appCtx.dbCnx.Exec("INSERT INTO packs_samples (directory, name) VALUES ((?), (?))", dir, base); err != nil {
					return err
				}
			}
			return nil
		})
	return err
}

func isSampleInPacks(sample_filename, packName string, dirs []string) (bool, error) {
	var count int
	dir := appCtx.libRootPath + "/" + packName

	for _, d := range dirs { // Should loop in reverse, would exec less queries
		dir += "/" + d
		stmt := fmt.Sprintf("SELECT COUNT(*) FROM packs_samples WHERE directory = '%s' AND name = '%s'", dir, sample_filename)
		_ = appCtx.dbCnx.QueryRow(stmt).Scan(&count)
		if count > 0 {
			return true, nil
		}
	}
	return false, nil
}
