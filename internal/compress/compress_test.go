package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"
	"testing"
)

func TestNewGzipReader(t *testing.T) {
	input := "CREATE TABLE test (id INT); INSERT INTO test VALUES (1);"
	reader := NewGzipReader(strings.NewReader(input))

	compressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("reading compressed data: %v", err)
	}

	// Decompress and verify
	gr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("creating gzip reader: %v", err)
	}
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("decompressing: %v", err)
	}

	if string(decompressed) != input {
		t.Errorf("got %q, want %q", string(decompressed), input)
	}
}

func TestNewGzipReader_Large(t *testing.T) {
	input := strings.Repeat("INSERT INTO test VALUES (1);\n", 10000)
	reader := NewGzipReader(strings.NewReader(input))

	compressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("reading compressed data: %v", err)
	}

	if len(compressed) >= len(input) {
		t.Error("compressed data should be smaller than input")
	}

	gr, _ := gzip.NewReader(bytes.NewReader(compressed))
	decompressed, _ := io.ReadAll(gr)
	if string(decompressed) != input {
		t.Error("round-trip failed for large input")
	}
}
