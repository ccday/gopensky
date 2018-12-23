package gopensky

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDeserializeStates(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "sample.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var raw map[string]interface{}
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		t.Fatal(err)
	}

	deserializeStates(raw["states"].([]interface{}))
}
