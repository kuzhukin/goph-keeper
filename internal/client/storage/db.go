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
	UserStorage
	DataStorage
	WalletStorage
	Stop() error
}

type DataStorage interface {
	SaveData(u *User, r *Record) (uint64, error)
	LoadData(u *User, name string) (*Record, error)
	ListData(u *User) ([]*Record, error)
	DeleteData(u *User, r *Record) error
}

type UserStorage interface {
	Register(login string, password string, cryptokey string) error
	User() (*User, error)
}

type WalletStorage interface {
	CreateCard(u *User, c *BankCard) error
	ListCard(u *User) ([]*BankCard, error)
	DeleteCard(u *User, cardNumber string) error
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

func (s *DbStorage) Register(login string, password string, cryptoKey string) error {
	cryptoKeyBase64 := base64.RawStdEncoding.EncodeToString([]byte(cryptoKey))

	if err := s.changeCurrentUserStatus(); err != nil {
		return err
	}

	fmt.Println("User was registred. Context was switched to a new user.")

	return s.addNewUser(login, password, cryptoKeyBase64)
}

func (s *DbStorage) addNewUser(login string, password string, cryptoKey string) error {
	query := prepareInsertUserQuery(login, password, cryptoKey)

	_, err := s.db.Exec(query.request, query.args...)
	if err != nil {
		if isUniqueConstraint(err) {
			fmt.Println("User alreay registred in client")

			return nil
		}

		return err
	}

	return nil
}

func (s *DbStorage) changeCurrentUserStatus() error {
	var err error

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		panicErr := recover()
		if err != nil {
			fmt.Printf("db transaction panics with error: %v\n", panicErr)

			err = tx.Rollback()
		}
	}()

	q := prepareGetUserQuery()
	rows, err := tx.Query(q.request, q.args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		// don't have active user
		return nil
	}

	u := User{IsActive: true}
	if err = rows.Scan(&u.Login, &u.Password, &u.CryptoKey); err != nil {
		return err
	}

	if err = rows.Err(); err != nil {
		return err
	}

	q = prepareChangeActiveQuery(u.Login)
	_, err = tx.Exec(q.request, q.args...)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func isUniqueConstraint(err error) bool {
	sqliteErr, ok := err.(sqlite3.Error)
	if !ok {
		return false
	}

	return sqliteErr.Code.Error() == sqlite3.ErrConstraintUnique.Error()
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

func (s *DbStorage) DeleteData(u *User, r *Record) error {
	q := prepareDeleteDataQuery(u.Login, r.Name)

	_, err := s.db.Exec(q.request, q.args...)

	return err
}

func (s *DbStorage) SaveData(u *User, r *Record) (uint64, error) {
	rev, err := s.getRevision(u, r)
	if err != nil {
		if errors.Is(err, ErrDataNotExist) {
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

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (s *DbStorage) updateData(u *User, r *Record) error {
	q := prepareUpdateDataQuery(u.Login, r.Name, r.Data)

	res, err := s.db.Exec(q.request, q.args...)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
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

func (s *DbStorage) LoadData(u *User, name string) (*Record, error) {
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

func (s *DbStorage) ListData(u *User) ([]*Record, error) {
	query := prepareListDataQuery(u.Login)

	rows, err := s.db.Query(query.request, query.args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	records := []*Record{}

	for rows.Next() {
		r := Record{}
		err = rows.Scan(&r.Name, &r.Data, &r.Revision)
		if err != nil {
			return nil, err
		}

		records = append(records, &r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// TODO:
func (s *DbStorage) CreateCard(u *User, c *BankCard) error {
	return nil
}

func (s *DbStorage) DeleteCard(u *User, number string) error {
	return nil
}

func (s *DbStorage) ListCard(u *User) ([]*BankCard, error) {
	return []*BankCard{}, nil
}

func (s *DbStorage) Stop() error {
	return s.db.Close()
}
