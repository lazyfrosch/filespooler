package sender

import (
	"encoding/gob"
	"io"
)

type FileData struct {
	RawName    string
	RawContent []byte
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
	return f.RawName
}

func (f *FileData) Content() []byte {
	return f.RawContent
}

func (f *FileData) SetContent(content []byte) {
	f.RawContent = content
}

func (f *FileData) Size() int {
	return len(f.RawContent)
}
