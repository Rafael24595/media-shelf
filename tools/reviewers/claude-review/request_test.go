package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRequestKeepsUnknownFieldsInAlphabeticalOrder(t *testing.T) {
	in := `{"contract":"media-shelf-request/1","id":"REQ-1","zeta":{"nested":[1,2]},"status":"pending","alpha":null}`
	r, err := parseRequest([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	if err := setAll(r, "status", "in_progress", "claimed_by", "x"); err != nil {
		t.Fatal(err)
	}
	out, err := r.encode()
	if err != nil {
		t.Fatal(err)
	}
	want := `{
  "alpha": null,
  "claimed_by": "x",
  "contract": "media-shelf-request/1",
  "id": "REQ-1",
  "status": "in_progress",
  "zeta": {
    "nested": [
      1,
      2
    ]
  }
}
`
	if string(out) != want {
		t.Errorf("encoded:\n%s\nwant:\n%s", out, want)
	}
}

func TestRequestReadsTextFields(t *testing.T) {
	r, _ := parseRequest([]byte(`{"id":"REQ-1","n":3}`))
	if r.text("id") != "REQ-1" || r.text("n") != "" || r.text("missing") != "" {
		t.Errorf("text: %q %q %q", r.text("id"), r.text("n"), r.text("missing"))
	}
}

func TestRequestRejectsWhatIsNotAnObject(t *testing.T) {
	for _, in := range []string{`[]`, `"x"`, ``, `null`, `{"a":`, `{"a":1} trailing`, `{"a":1}{}`} {
		if _, err := parseRequest([]byte(in)); err == nil {
			t.Errorf("%q: no error", in)
		}
	}
}

func TestWriteRequestReplacesTheFileWithoutLeftovers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "REQ-1.json")
	if err := os.WriteFile(path, []byte(`{"id":"REQ-1"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	r, _ := readRequest(path)
	_ = r.set("status", "failed")
	if err := writeRequest(path, r); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `"status": "failed"`) || !strings.HasSuffix(string(data), "}\n") {
		t.Errorf("written: %s", data)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Errorf("files left: %v", entries)
	}
}
