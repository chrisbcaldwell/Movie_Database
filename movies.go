package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

const (
	dataFilePath string = "./data"
	dbPath       string = "./movies.db"
	filePrefix   string = "IMDB-"
)

func main() {
	// os.Create(dbPath) // create empty db file
	files, err := os.ReadDir(dataFilePath)
	if err != nil {
		fmt.Println("path not found:", dataFilePath)
		log.Fatal(err)
	}

	// set up the table schemas
	// schema map keys should exactly match file names with "IMDB-" and ".csv" removed
	schemas := make(map[string]table)
	schemas["directors"] = table{
		name: "directors",
		fields: []field{
			newField("director_id", "INTEGER", true),
			newField("first_name", "TEXT"),
			newField("last_name", "TEXT"),
		},
	}
	schemas["movies_genres"] = table{
		name: "movies_genres",
		fields: []field{
			newField("movie_id", "INTEGER", false, true),
			newField("genre", "TEXT"),
		},
		fkeys: []fkey{{name: "movie_id", ref: "movies"}},
	}
	schemas["roles"] = table{
		name: "roles",
		fields: []field{
			newField("actor_id", "INTEGER", false, true),
			newField("movie_id", "INTEGER", false, true),
			newField("role", "TEXT"),
		},
		fkeys: []fkey{
			{name: "actor_id", ref: "actors"},
			{name: "movie_id", ref: "movies"},
		},
	}
	schemas["movies"] = table{
		name: "movies",
		fields: []field{
			newField("movie_id", "INTEGER", true),
			newField("name", "TEXT"),
			newField("year", "INTEGER"),
			newField("rank", "DOUBLE"),
		},
	}
	schemas["directors_genres"] = table{
		name: "directors_genres",
		fields: []field{
			newField("director_id", "INTEGER", false, true),
			newField("genre", "TEXT"),
			newField("prob", "DOUBLE"),
		},
		fkeys: []fkey{{name: "director_id", ref: "directors"}},
	}
	schemas["actors"] = table{
		name: "actors",
		fields: []field{
			newField("actor_id", "INTEGER", true),
			newField("first_name", "TEXT"),
			newField("last_name", "TEXT"),
			newField("gender", "TEXT"),
		},
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Println("Unable to open or create database", dbPath)
		log.Fatal(err)
	}

	for _, file := range files {
		schema := strings.Replace(file.Name(), filePrefix, "", -1)
		schema = strings.Replace(schema, ".csv", "", -1)
		updateTableFromCSV(db, file, schemas[schema])
	}

	// looking at the status of the tables
	for _, table := range schemas {
		q := "SELECT * FROM " + table.name
		result, err := db.Query(q)
		if err != nil {
			fmt.Println("Error querying table", table.name)
			log.Fatal(err)
		}
		fmt.Println("Status of table " + table.name + ":")
		fmt.Println("Column names and data types:")
		fmt.Println(result.Columns())
		fmt.Println(result.ColumnTypes())
		var count int
		for result.Next() {
			count += 1
		}
		fmt.Println("Number of rows:", count)

	}

}

type field struct {
	name string
	typ  string
	pkey bool
	fkey bool
}

type fkey struct {
	name string
	ref  string
}

type table struct {
	name   string
	fields []field
	fkeys  []fkey
}

// newField creates a new entity of type field.
// name and type are required.
// type needs to be the type name recognized by SQLite
// if no key booleans are passed both primary and foreign keys are left as false
// if one key is passed it is assigned to pkey
// to assign fkey it must be the second key boolean
// example: create a field as foreign key:
// f := newField("fname", "TYPE", false, true)
// keys booleans byond the second one are ignored
func newField(n string, t string, keys ...bool) field {
	f := field{
		name: n,
		typ:  t,
	}

	nkeys := len(keys)
	switch {
	case nkeys >= 2:
		f.pkey = keys[0]
		f.fkey = keys[1]
	case nkeys == 1:
		f.pkey = keys[0]
	}

	return f
}

// updateTableFromCSV creates a new table if needed, then updates based
// on the schema given by schema
func updateTableFromCSV(db *sql.DB, file os.DirEntry, schema table) error {
	path := dataFilePath + "/" + file.Name()
	data, _, err := readCsv(path) // read CSV ignoring the header
	if err != nil {
		log.Fatal(err)
	}
	addToTable(db, schema, data)
	return nil
}

func readCsv(filePath string) ([][]string, []string, error) {
	// readCsv opens and reads a CSV file and returns:
	// * Records: a slice of slices of strings, one slice per row of the input
	// file, then one string per field in the row
	// * a slice of strings, each string is one field name from the header row
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Unable to read input file " + filePath)
		return nil, nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	headerRow, err := r.Read()
	if err != nil {
		fmt.Println("Unable to parse input file as CSV for " + filePath)
		return nil, nil, err
	}

	records, err := r.ReadAll()
	if err != nil {
		fmt.Println("Unable to parse input file as CSV for " + filePath)
		return nil, nil, err
	}

	return records, headerRow, nil
}

func addToTable(db *sql.DB, t table, data [][]string) error {
	// start the CREATE TABLE query
	createQuery := "CREATE TABLE IF NOT EXISTS " + t.name + " (\n"
	// add the basic info for each field
	for _, field := range t.fields {
		line := field.name + " " + field.typ
		if field.pkey {
			line = line + " PRIMARY KEY,\n"
		} else {
			line = line + ",\n"
		}
		createQuery += line
	}
	// add foreign key details
	for _, fkey := range t.fkeys {
		line := "FOREIGN KEY (" + fkey.name + ")\n"
		line += "REFERENCES " + fkey.ref + " (" + fkey.name + "),\n"
		createQuery += line
	}
	createQuery = strings.TrimSuffix(createQuery, ",\n") // remove the last comma
	createQuery += "\n);"

	/////////////
	/////////////
	fmt.Println(createQuery)
	/////////////
	/////////////

	_, err := db.Exec(createQuery)
	if err != nil {
		fmt.Println("Error creating table", t.name)
		log.Fatal(err)
	}

	// insert data row by row
	totalRows := len(data)

	for i, row := range data {
		fmt.Println("Table", t.name, "inserting row", i, "of", totalRows)
		addDataRow(row, db, t)
	}

	return nil

}

func addDataRow(row []string, db *sql.DB, t table) error {
	// SQLite takes text with single quotes aorund it.
	// numerical fields are cast properly if possible
	// single quotes w/in text need to be '' as escape characters
	for j := range row {
		row[j] = strings.Replace(row[j], "'", "''", -1)
		row[j] = "'" + row[j] + "'"
	}
	addQuery := "INSERT INTO " + t.name + " VALUES (" + strings.Join(row, ", ") + ");"

	_, err := db.Exec(addQuery)
	if err != nil {
		fmt.Println("Error adding rows to table", t.name)
		log.Fatal(err)
	}
	return nil
}
