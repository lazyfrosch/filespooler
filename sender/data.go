package sender

import (
	"encoding/gob"
	"io"
)

type FileData struct {
	name    string
	content []byte
}

func NewFileData(name string) *FileData {
	return &FileData{name, nil}
}

func DecodeGobFileData(reader io.Reader) (*FileData, error) {
	file := new(FileData)
	dec := gob.NewDecoder(reader)
	if err := dec.Decode(file); err != nil {
		return nil, err
	}
	return file, nil
}

func (f *FileData) Name() string {
	return f.name
}

func (f *FileData) Content() []byte {
	return f.content
}

func (f *FileData) SetContent(content []byte) {
	f.content = content
}

func (f *FileData) Size() int {
	return len(f.content)
}
