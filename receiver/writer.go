package receiver

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type FileWriter struct {
	Path string
}

type FileData struct {
	Name string
	Content []byte
}

func NewFileWriter(path string) (*FileWriter, error) {
	w := FileWriter{ Path: path }
	err := w.init()
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (w FileWriter) init() error {
	info, err := os.Stat(w.Path)
	if os.IsNotExist(err) {
		err = os.Mkdir(w.Path, 0755)
		if err != nil {
			return err
		}

		return nil
	} else if err != nil {
		return err
	}

	if ! info.IsDir() {
		return fmt.Errorf("target path %s exists and is not a directory", w.Path)
	}

	return nil
}

func (w FileWriter) WriteFile(f *FileData) error {
	filePath := path.Join(w.Path, f.Name)
	return ioutil.WriteFile(filePath, f.Content, 0644)
}