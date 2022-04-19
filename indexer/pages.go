package indexer

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

type Pages struct {
	db          *sql.DB
	initialized bool
}

func (p *Pages) Init() {
	dbName := "pages.db"
	exists, err := p.fileExists(dbName)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		log.Println("Creating SQLite database for pages...")
		file, err := os.Create("pages.db")
		if err != nil {
			log.Fatal(err)
		}
		err = file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}

	p.db, err = sql.Open("sqlite3", dbName)
	if err != nil {
		log.Fatal(err)
	}

	p.initialized = true
	p.createTable()
}

func (p *Pages) createTable() {
	if !p.initialized {
		log.Fatal("Must initialize database connection before operating on it")
	}
	log.Println("Creating table pages...")
	createDB := `CREATE TABLE IF NOT EXISTS pages (
					id INTEGER PRIMARY KEY,
					url TEXT UNIQUE
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_url ON pages (url);`
	statement, err := p.db.Prepare(createDB)
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Pages) CheckExists(url string) bool {
	if !p.initialized {
		log.Fatal("Must initialize database connection before operating on it")
	}
	var exists bool
	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM pages WHERE url = '%s');`, url)
	result := p.db.QueryRow(query)
	err := result.Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}
	return exists
}

func (p *Pages) InsertPage(url string) {
	if !p.initialized {
		log.Fatal("Must initialize database connection before operating on it")
	}
	query := fmt.Sprintf(`INSERT INTO pages (url) VALUES ('%s');`, url)
	statement, err := p.db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
	return
}

func (p *Pages) fileExists(path string) (bool, error) {
	// Check if file exists on disk
	// Taken from: https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}
