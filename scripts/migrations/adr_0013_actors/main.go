// Command adr_0013_actors migrates a data folder from schema ADR-0008 to ADR-0013: it adds
// the `actors` column at the end of movies.csv, video_games.csv and series.csv, moves the
// tags that named actors to it (actors.go), and sets schema_version in shelf.json.
//
//	go run . <data folder>
//
// The column is added keeping every byte (a comma before the line break that ends each
// record); the tags are then moved and the file encoded as the data tools do. Each step is
// checked against the previous one (only tags and actors may differ), and nothing is
// written unless every file can be migrated.
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	from   = "ADR-0008"
	to     = "ADR-0013"
	column = "actors"
)

var files = []string{"movies.csv", "video_games.csv", "series.csv"}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: adr_0013_actors <data folder>")
		os.Exit(2)
	}
	if err := migrate(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Println("migrated to", to)
}

func migrate(dir string) error {
	manifest := filepath.Join(dir, "shelf.json")
	raw, err := os.ReadFile(manifest)
	if err != nil {
		return err
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("shelf.json: %w", err)
	}
	if m["schema_version"] != from {
		return fmt.Errorf("shelf.json: schema_version is %v, this migration starts from %s", m["schema_version"], from)
	}
	out := map[string][]byte{}
	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		migrated, err := addColumn(data, column)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		migrated, moved, err := moveActorTags(migrated)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		fmt.Printf("%s: actor tags moved to actors in %d row(s)\n", name, moved)
		out[name] = migrated
	}
	m["schema_version"] = to
	manifestOut, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	out["shelf.json"] = append(manifestOut, '\n')
	for _, name := range append(files, "shelf.json") {
		if err := os.WriteFile(filepath.Join(dir, name), out[name], 0o644); err != nil {
			return err
		}
	}
	return nil
}

// addColumn appends an empty column named name to every record of a CSV file (LF line
// endings, the data convention) and checks the result.
func addColumn(data []byte, name string) ([]byte, error) {
	if bytes.Contains(data, []byte("\r")) {
		return nil, errors.New("CR found: the data uses LF line endings")
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		return nil, errors.New("the file must end with a line break")
	}
	header, _, _ := bytes.Cut(data, []byte("\n"))
	for _, h := range bytes.Split(header, []byte(",")) {
		if string(h) == name {
			return nil, fmt.Errorf("the %s column already exists", name)
		}
	}
	var out bytes.Buffer
	out.Grow(len(data) + 1024)
	quoted, first := false, true
	for _, b := range data {
		if b == '"' {
			quoted = !quoted // an escaped quote ("") toggles twice
		}
		if b == '\n' && !quoted {
			if first {
				out.WriteString("," + name)
				first = false
			} else {
				out.WriteByte(',')
			}
		}
		out.WriteByte(b)
	}
	if quoted {
		return nil, errors.New("unterminated quoted field")
	}
	return out.Bytes(), check(data, out.Bytes())
}

// check parses both versions and verifies that every record kept its fields and gained one
// empty field.
func check(before, after []byte) error {
	b, err := csv.NewReader(bytes.NewReader(before)).ReadAll()
	if err != nil {
		return fmt.Errorf("reading the original: %w", err)
	}
	a, err := csv.NewReader(bytes.NewReader(after)).ReadAll()
	if err != nil {
		return fmt.Errorf("reading the result: %w", err)
	}
	if len(a) != len(b) {
		return fmt.Errorf("%d records became %d", len(b), len(a))
	}
	for i := range b {
		if len(a[i]) != len(b[i])+1 {
			return fmt.Errorf("record %d: %d fields became %d", i+1, len(b[i]), len(a[i]))
		}
		for j := range b[i] {
			if a[i][j] != b[i][j] {
				return fmt.Errorf("record %d, field %d changed", i+1, j+1)
			}
		}
		if i > 0 && a[i][len(b[i])] != "" {
			return fmt.Errorf("record %d: the new field is not empty", i+1)
		}
	}
	return nil
}
