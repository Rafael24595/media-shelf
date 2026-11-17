package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddColumnKeepsQuotedLineBreaks(t *testing.T) {
	in := "id,notes\nMOV-0001,\"two\nlines, \"\"quoted\"\"\"\nMOV-0002,plain\n"
	got, err := addColumn([]byte(in), "actors")
	if err != nil {
		t.Fatal(err)
	}
	want := "id,notes,actors\nMOV-0001,\"two\nlines, \"\"quoted\"\"\",\nMOV-0002,plain,\n"
	if string(got) != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

func TestAddColumnRefuses(t *testing.T) {
	cases := map[string]string{
		"id,actors\nMOV-0001,x\n": "already exists",
		"id\r\nMOV-0001\r\n":      "CR found",
		"id\nMOV-0001":            "must end with a line break",
		"id,n\nMOV-0001,\"open\n": "unterminated",
	}
	for in, want := range cases {
		if _, err := addColumn([]byte(in), "actors"); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: error %v, want %q", in, err, want)
		}
	}
}

func TestMigrate(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("shelf.json", `{"schema_version": "ADR-0008"}`)
	write("movies.csv", "id,tags,reviewed\nMOV-0001,tom-hanks|drama,false\n")
	write("video_games.csv", "id,tags,reviewed,platform\n")
	write("series.csv", "id,tags,reviewed,series_type,seasons\nSER-0001,,false,series,1|2\n")
	write("books.csv", "id,isbn\n")
	if err := migrate(dir); err != nil {
		t.Fatal(err)
	}
	read := func(name string) string {
		b, _ := os.ReadFile(filepath.Join(dir, name))
		return string(b)
	}
	if read("movies.csv") != "id,tags,reviewed,actors\nMOV-0001,drama,false,Tom Hanks\n" || read("video_games.csv") != "id,tags,reviewed,platform,actors\n" ||
		read("series.csv") != "id,tags,reviewed,series_type,seasons,actors\nSER-0001,,false,series,1|2,\n" {
		t.Errorf("files: %q %q %q", read("movies.csv"), read("video_games.csv"), read("series.csv"))
	}
	if read("books.csv") != "id,isbn\n" {
		t.Error("books.csv must not change")
	}
	if !strings.Contains(read("shelf.json"), `"schema_version": "ADR-0013"`) {
		t.Errorf("shelf.json: %s", read("shelf.json"))
	}
	if err := migrate(dir); err == nil || !strings.Contains(err.Error(), "starts from ADR-0008") {
		t.Errorf("a second run must refuse: %v", err)
	}
}

func TestMigrateWritesNothingWhenAFileFails(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "shelf.json"), []byte(`{"schema_version": "ADR-0008"}`), 0o644)
	os.WriteFile(filepath.Join(dir, "movies.csv"), []byte("id,tags\nMOV-0001,\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "video_games.csv"), []byte("id,tags,actors\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "series.csv"), []byte("id,tags\n"), 0o644)
	if err := migrate(dir); err == nil {
		t.Fatal("expected an error")
	}
	b, _ := os.ReadFile(filepath.Join(dir, "movies.csv"))
	if string(b) != "id,tags\nMOV-0001,\n" {
		t.Error("movies.csv was written")
	}
}

func TestMoveActorTags(t *testing.T) {
	in := "id,tags,notes,actors\n" +
		"MOV-0138,al-pacino|heist|robert-de-niro,\"a, b\",\n" +
		"MOV-0714,jerry-lewis|dean-martin,,\n" +
		"MOV-0001,quentin-tarantino|crime,,\n" +
		"MOV-0002,van-damme,,Jean-Claude Van Damme|Someone\n"
	got, n, err := moveActorTags([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	want := "id,tags,notes,actors\n" +
		"MOV-0138,heist,\"a, b\",Al Pacino|Robert De Niro\n" +
		"MOV-0714,,,Dean Martin|Jerry Lewis\n" +
		"MOV-0001,quentin-tarantino|crime,,\n" +
		"MOV-0002,,,Jean-Claude Van Damme|Someone\n"
	if string(got) != want || n != 3 {
		t.Fatalf("%d rows; got\n%s\nwant\n%s", n, got, want)
	}
}
