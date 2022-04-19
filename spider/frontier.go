package spider

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

type Frontier struct {
	db          *sql.DB
	initialized bool
}

func (f *Frontier) Init() {
	dbName := "frontier.db"
	exists, err := f.fileExists(dbName)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		log.Println("Creating SQLite database for frontier...")
		file, err := os.Create("frontier.db")
		if err != nil {
			log.Fatal(err)
		}
		err = file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}

	f.db, err = sql.Open("sqlite3", dbName)
	if err != nil {
		log.Fatal(err)
	}

	f.initialized = true
	f.createTable()
}

func (f *Frontier) createTable() {
	if !f.initialized {
		log.Fatal("Must initialize database connection before operating on it")
	}
	log.Println("Creating table frontier...")
	createDB := `CREATE TABLE IF NOT EXISTS frontier (
					url TEXT PRIMARY KEY
				);`
	statement, err := f.db.Prepare(createDB)
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
}

func (f *Frontier) PopURL() string {
	// Query and return a URL from the frontier DB
	// true if frontier isn't empty, false otherwise
	if !f.initialized {
		log.Fatal("Must initialize database connection before operating on it")
	}
	var url string
	query := `SELECT url FROM frontier LIMIT 1`
	result := f.db.QueryRow(query)
	err := result.Scan(&url)
	if err != nil {
		return ""
	}
	query = fmt.Sprintf(`DELETE FROM frontier WHERE url = '%s';`, url)
	statement, err := f.db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
	return url
}

func (f *Frontier) CheckURLInFrontier(url string) bool {
	if !f.initialized {
		log.Fatal("Must initialize database connection before operating on it")
	}
	var exists bool
	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM frontier WHERE url = '%s');`, url)
	result := f.db.QueryRow(query)
	err := result.Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}
	return exists
}

func (f *Frontier) InsertPage(url string) {
	if !f.initialized {
		log.Fatal("Must initialize database connection before operating on it")
	}
	query := fmt.Sprintf(`INSERT INTO frontier (url) VALUES ('%s');`, url)
	statement, err := f.db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
	return
}

func (f *Frontier) fileExists(path string) (bool, error) {
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
