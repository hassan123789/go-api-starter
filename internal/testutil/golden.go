// Package testutil provides utilities for golden file testing.
package testutil

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

// Dir returns the path to the testdata/golden directory.
func Dir() string {
	return filepath.Join("testdata", "golden")
}

// Path returns the full path for a golden file.
func Path(name string) string {
	return filepath.Join(Dir(), name+".golden.json")
}

// Get reads and returns the content of a golden file.
func Get(t *testing.T, name string) []byte {
	t.Helper()

	path := Path(name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("golden file not found: %s\nRun with -update flag to create it", path)
		}
		t.Fatalf("failed to read golden file: %v", err)
	}

	return data
}

// Update writes data to a golden file.
func Update(t *testing.T, name string, data []byte) {
	t.Helper()

	path := Path(name)

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create golden directory: %v", err)
	}

	// Format JSON for readability
	var formatted []byte
	if json.Valid(data) {
		var v interface{}
		if err := json.Unmarshal(data, &v); err == nil {
			formatted, _ = json.MarshalIndent(v, "", "  ")
		}
	}
	if formatted == nil {
		formatted = data
	}

	if err := os.WriteFile(path, formatted, 0644); err != nil {
		t.Fatalf("failed to write golden file: %v", err)
	}

	t.Logf("updated golden file: %s", path)
}

// Assert compares actual data against a golden file.
// If -update flag is provided, it updates the golden file instead.
func Assert(t *testing.T, name string, actual []byte) {
	t.Helper()

	if *update {
		Update(t, name, actual)
		return
	}

	expected := Get(t, name)

	// For JSON, compare parsed structures to ignore formatting differences
	if json.Valid(actual) && json.Valid(expected) {
		var actualJSON, expectedJSON interface{}
		require.NoError(t, json.Unmarshal(actual, &actualJSON))
		require.NoError(t, json.Unmarshal(expected, &expectedJSON))
		assert.Equal(t, expectedJSON, actualJSON, "golden file mismatch for %s", name)
	} else {
		assert.Equal(t, string(expected), string(actual), "golden file mismatch for %s", name)
	}
}

// AssertJSON marshals v to JSON and compares against a golden file.
func AssertJSON(t *testing.T, name string, v interface{}) {
	t.Helper()

	data, err := json.Marshal(v)
	require.NoError(t, err)

	Assert(t, name, data)
}

// AssertJSONResponse is specifically for HTTP response bodies.
// It normalizes timestamps and IDs for reproducible tests.
func AssertJSONResponse(t *testing.T, name string, body []byte, normalizers ...Normalizer) {
	t.Helper()

	normalized := NormalizeJSON(t, body, normalizers...)
	Assert(t, name, normalized)
}

// Normalizer is a function that normalizes JSON for comparison.
type Normalizer func(map[string]interface{})

// NormalizeTimestamps replaces timestamp fields with a placeholder.
func NormalizeTimestamps(fields ...string) Normalizer {
	return func(m map[string]interface{}) {
		for _, field := range fields {
			if _, ok := m[field]; ok {
				m[field] = "<<TIMESTAMP>>"
			}
		}
	}
}

// NormalizeIDs replaces ID fields with a placeholder.
func NormalizeIDs(fields ...string) Normalizer {
	return func(m map[string]interface{}) {
		for _, field := range fields {
			if _, ok := m[field]; ok {
				m[field] = "<<ID>>"
			}
		}
	}
}

// NormalizeJSON applies normalizers to JSON data.
func NormalizeJSON(t *testing.T, data []byte, normalizers ...Normalizer) []byte {
	t.Helper()

	var v interface{}
	require.NoError(t, json.Unmarshal(data, &v))

	normalizeValue(v, normalizers)

	result, err := json.Marshal(v)
	require.NoError(t, err)

	return result
}

func normalizeValue(v interface{}, normalizers []Normalizer) {
	switch val := v.(type) {
	case map[string]interface{}:
		for _, n := range normalizers {
			n(val)
		}
		for _, child := range val {
			normalizeValue(child, normalizers)
		}
	case []interface{}:
		for _, item := range val {
			normalizeValue(item, normalizers)
		}
	}
}
