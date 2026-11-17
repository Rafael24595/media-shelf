package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// request is a request file (docs/review-requests.md) as its fields, each kept as raw JSON
// until it is read or set: the contract asks tools to keep the fields they do not know. It
// is written with its fields in alphabetical order, as media-shelf-ui writes it, so both
// tools leave the same order.
type request map[string]json.RawMessage

func readRequest(path string) (request, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseRequest(data)
}

// parseRequest reads a JSON object; anything else, or anything after it, is an error.
func parseRequest(data []byte) (request, error) {
	var r request
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if r == nil {
		return nil, fmt.Errorf("a request is a JSON object")
	}
	return r, nil
}

// text is the value of a text field, or "" when it is missing or not text.
func (r request) text(name string) string {
	var s string
	if json.Unmarshal(r[name], &s) != nil {
		return ""
	}
	return s
}

// set sets a field to any value that encodes as JSON.
func (r request) set(name string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	r[name] = raw
	return nil
}

// encode writes the request indented with two spaces and a final newline (UTF-8, LF).
func (r request) encode() ([]byte, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// writeRequest replaces the file at once: written next to it, then renamed over it, so a
// reader never sees half a request.
func writeRequest(path string, r request) error {
	data, err := r.encode()
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
