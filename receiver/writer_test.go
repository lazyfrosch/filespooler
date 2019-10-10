package receiver

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func getTempDir() string {
	temp := os.Getenv("TEMP")
	if temp == "" {
		temp = "/tmp"
	}

	tempPath, err := ioutil.TempDir(temp, "filespool")
	if err != nil {
		panic(err)
	}

	return tempPath
}

func cleanupTempDir(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		panic(err)
	}
}

func TestFileWriter(t *testing.T) {
	tempPath := getTempDir()
	defer cleanupTempDir(tempPath)

	w, err := NewFileWriter(tempPath)
	if err != nil {
		t.Fatal(err)
	}

	testWrite(t, w, tempPath, "test1", []byte("abcdef"))
}

func testWrite(t *testing.T, w *FileWriter, tempPath string, name string, content []byte) {
	data := FileData{Name: name, Content: content}
	filePath := path.Join(tempPath, name)

	err := w.WriteFile(&data)
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(filePath)
	if err != nil {
		t.Fatal("could not stat written file", err)
	}

	writtenContent, err := ioutil.ReadFile(filePath)
	if err != nil || bytes.Compare(content, writtenContent) != 0 {
		t.Fatal("content is not identical")
	}
}
