package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Record struct {
	Callsign  string `json:"callsign"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Address   string `json:"address"`
	City      string `json:"city"`
	State     string `json:"state"`
}

const (
	databaseFile = "mydatabase.db"
	tableName    = "mytable"
	appname      = "com.ki4hdu.qrz"
)

var debug bool

func getCacheDir() (string, error) {
	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	// Determine the cache directory based on the operating system
	var cacheDir string
	switch os := runtime.GOOS; os {
	case "darwin":
		cacheDir = filepath.Join(currentUser.HomeDir, "Library", "Caches", appname)
	case "linux":
		cacheDir = filepath.Join(currentUser.HomeDir, ".cache", appname)
	case "windows":
		cacheDir = filepath.Join(currentUser.HomeDir, "AppData", "Local", "Cache", appname)
	default:
		return "", fmt.Errorf("unsupported operating system: %s", os)
	}

	// Check if the directory already exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		// Directory doesn't exist, so create it
		err := os.MkdirAll(cacheDir, os.ModePerm)
		if err != nil {
			return "", err
		}
		if debug {
			log.Println("Directory created:", cacheDir)
		}
	} else if err != nil {
		// Some other error occurred
		return "", err
	}

	return cacheDir, nil
}

func createTable(db *sql.DB) error {
	query := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, tableName)
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	query = fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s (
				callsign TEXT PRIMARY KEY,
				firstname TEXT,
				lastname TEXT,
				address TEXT,
				city TEXT,
				state TEXT
				)
				`, tableName)

	_, err = db.Exec(query)
	return err
}

func insertData(db *sql.DB, callsign string, firstname string, lastname string, address string, city string, state string) (sql.Result, error) {
	query := fmt.Sprintf(`
				INSERT INTO %s (callsign, firstname, lastname, address, city, state) VALUES (?, ?, ?, ?, ?, ?)
				`, tableName)

	result, err := db.Exec(query, callsign, firstname, lastname, address, city, state)
	return result, err
}

func queryByRegex(db *sql.DB, regex string) ([]Record, error) {
	query := fmt.Sprintf(`
				SELECT callsign, firstname, lastname, address, city, state FROM %s WHERE callsign LIKE ?
				`, tableName)

	rows, err := db.Query(query, regex)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Record

	for rows.Next() {
		var record Record

		// Scan the values from the row into the Person struct
		err := rows.Scan(&record.Callsign, &record.Firstname, &record.Lastname, &record.Address, &record.City, &record.State)
		if err != nil {
			log.Fatal("Error scanning row:", err)
		}

		// Convert the struct to JSON
		jsonData, err := json.Marshal(record)
		if err != nil {
			log.Fatal("Error marshaling to JSON:", err)
		}

		// Print the JSON representation
		if debug {
			log.Println(string(jsonData))
		}
		results = append(results, record)
	}

	return results, nil
}

func createDatabase(db *sql.DB, input string) {
	// Create table if it doesn't exist
	if err := createTable(db); err != nil {
		log.Fatal(err)
	}

	// Read data from the file and insert into the database
	file, err := os.Open(input)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, "|")

		// Insert data into the database
		if _, err := insertData(db, fields[4], fields[8], fields[10], fields[15], fields[16], fields[17]); err != nil {
			log.Println("Error inserting data:", err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

}

func buildAppleMapsURL(record Record) string {
	return strings.ReplaceAll(
		fmt.Sprintf("http://maps.apple.com/?address=%s,%s,%s",
			record.Address,
			record.City,
			record.State),
		" ",
		"+")
}

func searchDatabase(db *sql.DB, call string) {

	// Query the table for rows matching a regex in column1
	results, err := queryByRegex(db, call)
	if err != nil {
		log.Fatal(err)
	}

	for _, result := range results {
		fmt.Printf("%s\n%s %s\n%s\n%s, %s\n%s\n",
			result.Callsign,
			result.Firstname,
			result.Lastname,
			result.Address,
			result.City,
			result.State,
			buildAppleMapsURL(result))
	}
}

func main() {
	var reloadDatabse bool
	var inputFile string

	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.BoolVar(&reloadDatabse, "r", false, "Reload database")
	flag.StringVar(&inputFile, "input", "/usr/local/share/xastir/fcc/EN.dat", "input file name")

	flag.Parse()

	otherArgs := flag.Args()

	cachedir, err := getCacheDir()
	if err != nil {
		log.Fatal(err)
	}
	fuilldbpath := filepath.Join(cachedir, databaseFile)

	// Open the SQLite database
	db, err := sql.Open("sqlite3", fuilldbpath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if reloadDatabse {
		if debug {
			log.Printf("Reloading database from %s\n", inputFile)
		}
		// createDatabase(db, inputFile)
	}

	if len(otherArgs) > 0 {
		for _, arg := range otherArgs {
			searchDatabase(db, arg)
		}
	}
}
