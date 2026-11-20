package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Rafael24595/go-assert/assert"
	"github.com/Rafael24595/go-assert/require"
)

func TestRequestKeepsUnknownFieldsInAlphabeticalOrder(t *testing.T) {
	in := `{"contract":"media-shelf-request/1","id":"REQ-1","zeta":{"nested":[1,2]},"status":"pending","alpha":null}`
	r, err := parseRequest([]byte(in))
	require.Nil(t, err)
	require.Nil(t, setAll(r, "status", "in_progress", "claimed_by", "x"))
	out, err := r.encode()
	require.Nil(t, err)
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
	assert.Equal(t, want, string(out))
}

func TestRequestReadsTextFields(t *testing.T) {
	r, _ := parseRequest([]byte(`{"id":"REQ-1","n":3}`))
	assert.Equal(t, "REQ-1", r.text("id"))
	assert.Equal(t, "", r.text("n"), "a number is not text")
	assert.Equal(t, "", r.text("missing"))
}

func TestRequestRejectsWhatIsNotAnObject(t *testing.T) {
	for _, in := range []string{`[]`, `"x"`, ``, `null`, `{"a":`, `{"a":1} trailing`, `{"a":1}{}`} {
		_, err := parseRequest([]byte(in))
		assert.NotNil(t, err, "%q", in)
	}
}

func TestWriteRequestReplacesTheFileWithoutLeftovers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "REQ-1.json")
	require.Nil(t, os.WriteFile(path, []byte(`{"id":"REQ-1"}`), 0o644))
	r, _ := readRequest(path)
	_ = r.set("status", "failed")
	require.Nil(t, writeRequest(path, r))
	data, _ := os.ReadFile(path)
	assert.Inside(t, `"status": "failed"`, string(data))
	assert.True(t, strings.HasSuffix(string(data), "}\n"), "ends with a new line: %s", data)
	entries, _ := os.ReadDir(dir)
	assert.Size(t, 1, entries, "files left: %v", entries)
}
