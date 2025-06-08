package main

import (
	//"encoding/csv"
	_ "modernc.org/sqlite"
	//"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	dataFilePath string = "./data"
	dbpath       string = "./movies.db"
	filePrefix   string = "IMDB-"
)

func main() {
	os.Create(dbpath) // create empty db file
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
	}
	schemas["roles"] = table{
		name: "roles",
		fields: []field{
			newField("actor_id", "INTEGER", false, true),
			newField("movie_id", "INTEGER", false, true),
			newField("role", "TEXT"),
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

	for _, file := range files {
		schema := strings.Replace(file.Name(), filePrefix, "", -1)
		schema = strings.Replace(schema, ".csv", "", -1)
		updateTableFromCSV(file, schemas[schema])
	}

}

type field struct {
	name string
	typ  string
	pkey bool
	fkey bool
}

type table struct {
	name   string
	fields []field
}

// updateTableFromCSV creates a new table if needed, then updates based
// on the schema given by schema
func updateTableFromCSV(file os.DirEntry, schema table) error {
	path := dataFilePath + "/" + file.Name()
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("Unable to read input file " + path)
		return err
	}
	defer f.Close()

	//temp
	fmt.Println(schema)

	return nil
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
