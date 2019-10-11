package sender

import (
	"bytes"
	"testing"
)

func TestNewFileData(t *testing.T) {
	f := NewFileData("test")
	if f.Name() != "test" {
		t.Fatal("Failed to create")
	}
}

func TestFileData_Content(t *testing.T) {
	data := []byte("testdata")
	f := NewFileData("test")
	f.SetContent(data)

	storedData := f.Content()
	if &storedData == &data {
		t.Fatal("No pointers should be used for data")
	}

	if bytes.Compare(data, storedData) != 0 {
		t.Fatal("Stored data does not match input")
	}
}

func TestFileData_Size(t *testing.T) {
	data := []byte("testdata")
	expected := 8
	f := NewFileData("test")

	if f.Size() != 0 {
		t.Fatal("Size should be 0 without data")
	}

	f.SetContent(data)
	if f.Size() != expected {
		t.Fatalf("Size should be %d without data", expected)
	}
}
