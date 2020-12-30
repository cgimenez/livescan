package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Ableton Packs dir already scanned ?
func isLivePacksDBScanned() (bool, error) {
	var cnx *sql.DB
	var err error

	if !fileExists("files/db.sqlite") {
		dbFile, err := os.Create("files/db.sqlite")
		if err != nil {
			return false, err
		}
		defer dbFile.Close()
	}

	if cnx, err = newDBCnx(); err != nil {
		return false, err
	}
	defer cnx.Close()

	var count int
	stmt := "SELECT COUNT(*) FROM sqlite_master WHERE name ='packs_samples' and type='table'"
	_ = cnx.QueryRow(stmt).Scan(&count)
	if count == 0 {
		stmt = "CREATE TABLE packs_samples (directory TEXT, name TEXT)"
		if _, err := cnx.Exec(stmt); err != nil {
			return false, err
		}
		return false, nil
	} else {
		stmt = "SELECT COUNT(*) FROM packs_samples"
		_ = cnx.QueryRow(stmt).Scan(&count)
		if count == 0 {
			return false, nil
		}
	}
	return true, nil
}

func scanLivePacks() error {
	var cnx *sql.DB
	var err error

	if appCtx.libRootPath == "" {
		return errors.New("libRootPath is not set")
	}

	if cnx, err = newDBCnx(); err != nil {
		return err
	}
	defer cnx.Close()

	if _, err = cnx.Exec("DELETE FROM packs_samples"); err != nil {
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
				if _, err = cnx.Exec("INSERT INTO packs_samples (directory, name) VALUES ((?), (?))", dir, base); err != nil {
					return err
				}
			}
			return nil
		})
	return err
}
