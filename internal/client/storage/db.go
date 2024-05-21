package storage

import (
	"database/sql"
	"errors"
)

var (
	ErrNotExist         = errors.New("doesn't exist")
	ErrUpdateConflict   = errors.New("updae conflict")
	ErrUserNotRegistred = errors.New("user isn't registred")
)

type Storage interface {
	Register(login string, password string) error
	User() (*User, error)
	Save(f *File) error
	Load(name string) (*File, error)
	Stop() error
}

type DbStorage struct {
	db *sql.DB
}

func StartNewDbStorage(dbName string) (*DbStorage, error) {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return nil, err
	}

	return &DbStorage{db: db}, nil
}

func (s *DbStorage) Register(login string, password string) error {
	return nil
}
func (s *DbStorage) User() (*User, error) {
	return nil, nil
}
func (s *DbStorage) Save(f *File) error {
	return nil
}

func (s *DbStorage) Load(name string) (*File, error) {
	return nil, nil
}

func (s *DbStorage) Stop() error {
	return s.db.Close()
}
