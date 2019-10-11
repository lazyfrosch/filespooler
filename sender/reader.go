package sender

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type FileReader struct {
	path string
}

func NewFileReader(path string) (*FileReader, error) {
	w := FileReader{path}
	stat, err := os.Stat(w.path)
	if err != nil {
		return nil, fmt.Errorf("could not stat source directory: %s", err.Error())
	}

	if !stat.IsDir() {
		return nil, fmt.Errorf("target path exists and is not a directory: %s", w.path)
	}

	return &w, nil
}

func (r FileReader) ReadDir() ([]*FileData, error) {
	files, err := ioutil.ReadDir(r.path)
	if err != nil {
		return nil, fmt.Errorf("could not open directory: %s", err)
	}

	var spool []*FileData
	for _, file := range files {
		name := file.Name()
		if file.IsDir() || name[0:1] == "." {
			continue
		}

		f, err := r.ReadFile(name)
		if err != nil {
			return nil, err
		}

		spool = append(spool, f)
	}

	return spool, nil
}

func (r FileReader) ReadFile(name string) (*FileData, error) {
	filePath := path.Join(r.path, name)
	f := NewFileData(name)

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read from file: %s", filePath)
	}

	f.SetContent(content)

	return f, nil
}

func (r FileReader) Delete(name string) error {
	filePath := path.Join(r.path, name)
	err := os.Remove(filePath)

	if err != nil {
		return fmt.Errorf("could not remove file %s: %s", filePath, err)
	}

	return nil
}
