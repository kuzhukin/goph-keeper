package storage

import (
	"os"
	"path/filepath"
)

type File struct {
	Name     string
	Data     string
	Revision uint64
}

func (f *File) Load() error {
	return nil
}

func (f *File) Save(outputDir string) error {
	_ = os.MkdirAll(outputDir, 0700)

	outputFilePath := filepath.Join(outputDir, f.Name)

	err := os.WriteFile(outputFilePath, []byte(f.Data), 0600)
	if err != nil {
		return err
	}

	return nil
}

func (f *File) Edit() error {
	return nil
}

type User struct {
	Login    string
	Password string
	IsActive bool
}
