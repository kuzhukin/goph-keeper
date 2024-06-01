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
)

var (
	ErrNotExist         = errors.New("doesn't exist")
	ErrUpdateConflict   = errors.New("updae conflict")
	ErrUserNotRegistred = errors.New("user isn't registred")
)

type Storage interface {
	Register(login string, password string, cryptokey string) error
	User() (*User, error)
	Save(u *User, r *Record) (uint64, error)
	Load(u *User, name string) (*Record, error)
	List(u *User) ([]*Record, error)
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
			if sqliteErr.Code.Error() == sqlite3.ErrConstraintUnique.Error() {
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

	cryptoKeyBase64 := ""

	err = rows.Scan(&user.Login, &user.Password, &cryptoKeyBase64)
	if err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	cryptoKey, err := base64.RawStdEncoding.DecodeString(cryptoKeyBase64)
	if err != nil {
		return nil, err
	}

	user.CryptoKey = cryptoKey

	return user, nil
}

func (s *DbStorage) Save(u *User, r *Record) (uint64, error) {
	rev, err := s.getRevision(u, r)
	if err != nil {
		if errors.Is(err, ErrNotExist) {
			return 1, s.saveNewData(u, r)
		}

		return 0, err
	}

	return rev, s.updateData(u, r)
}

func (s *DbStorage) saveNewData(u *User, r *Record) error {
	q := prepareAddDataQuery(u.Login, r.Name, r.Data)

	res, err := s.db.Exec(q.request, q.args...)
	if err != nil {
		return err
	}

	num, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if num != 1 {
		panic("affected more than one line in db. data isn't consistend")
	}

	return nil
}

func (s *DbStorage) updateData(u *User, r *Record) error {
	q := prepareUpdateDataQuery(u.Login, r.Name, r.Data)

	res, err := s.db.Exec(q.request, q.args...)
	if err != nil {
		return err
	}

	num, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if num != 1 {
		panic("affected more than one line in db. data isn't consistend")
	}

	return nil
}

var ErrDataNotExist = errors.New("data not exist")

func (s *DbStorage) getRevision(u *User, r *Record) (uint64, error) {
	query := preareGetRevisionQuery(u.Login, r.Name)

	rows, err := s.db.Query(query.request, query.args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, ErrDataNotExist
	}

	revision := uint64(0)
	err = rows.Scan(&revision)

	return revision, err
}

func (s *DbStorage) Load(u *User, name string) (*Record, error) {
	query := prepareGetDataQuery(u.Login, name)

	rows, err := s.db.Query(query.request, query.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrDataNotExist
	}

	record := &Record{Name: name}
	err = rows.Scan(&record.Data, &record.Revision)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (s *DbStorage) List(u *User) ([]*Record, error) {
	query := prepareListDataQuery(u.Login)

	rows, err := s.db.Query(query.request, query.args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	records := make([]*Record, 10)

	for rows.Next() {
		r := &Record{}
		err = rows.Scan(&r.Name, &r.Data, &r.Revision)
		if err != nil {
			return nil, err
		}

		records = append(records, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (s *DbStorage) Stop() error {
	return s.db.Close()
}
