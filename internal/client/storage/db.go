package storage

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kuzhukin/goph-keeper/internal/client/config"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrNotExist         = errors.New("doesn't exist")
	ErrUpdateConflict   = errors.New("updae conflict")
	ErrUserNotRegistred = errors.New("user isn't registred")
)

type Storage interface {
	Register(login string, password string, cryptokey string) error
	User() (*User, error)
	Save(f *File) error
	Load(name string) (*File, error)
	Stop() error
}

type DbStorage struct {
	db *sql.DB
}

func StartNewDbStorage(dbName string) (*DbStorage, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(homedir, config.DefaultAppDirName, dbName)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	storage := &DbStorage{db: db}

	if err = storage.initTables(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *DbStorage) initTables() error {
	var err error

	for _, t := range []string{
		createDataTableQuery,
		createUserTableQuery,
	} {
		if _, err = s.db.Exec(t); err != nil {
			fmt.Println("init table", t)

			return err
		}
	}

	return nil
}

const (
	ACTIVE  = 1
	PASSIVE = 0
)

func (s *DbStorage) Register(login string, password string, cryptoKey string) error {
	cryptoKeyBase64 := base64.RawStdEncoding.EncodeToString([]byte(cryptoKey))

	query := prepareInsertUserQuery(login, password, cryptoKeyBase64)

	_, err := s.db.Exec(query.request, query.args...)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if errors.Is(sqliteErr.Code, sqlite3.ErrNo(sqlite3.ErrConstraintUnique)) {
				fmt.Println("User alreay registred in client")

				return nil
			}
		}

		return err
	}

	fmt.Println("User was registred. Context was switched to a new user.")

	return nil
}

func (s *DbStorage) changeCurrentUserStatus(login string) error {
	// TODO: сделать текущего активного пользователя пассивным

	return nil
}

var ErrNotActiveOrRegistredUsers = errors.New("don't have active or registred users")

func (s *DbStorage) User() (*User, error) {
	query := prepareGetUserQuery()

	rows, err := s.db.Query(query.request)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrNotActiveOrRegistredUsers
	}

	user := &User{IsActive: true}
	err = rows.Scan(&user.Login, &user.Password)
	if err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return user, nil
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
