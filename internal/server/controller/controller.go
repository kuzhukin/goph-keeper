package controller

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kuzhukin/goph-keeper/internal/server/handler"
)

var _ handler.DataHolder = &Controller{}
var _ handler.Authenticator = &Controller{}
var _ handler.Registrator = &Controller{}

type Controller struct {
	db *sql.DB
}

func StartNewController(dataSourceName string) (*Controller, error) {
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("sql open dataSourceName=%s, err=%w", dataSourceName, err)
	}

	ctrl := &Controller{db: db}
	if err := ctrl.init(); err != nil {
		return nil, fmt.Errorf("init, err=%w", err)
	}

	return ctrl, nil
}

func (c *Controller) Stop() error {
	if err := c.db.Close(); err != nil {
		return fmt.Errorf("db close err=%w", err)
	}

	return nil
}

func (c *Controller) init() error {
	createTableQueries := []string{
		createDataTableQuery,
		createUsersTableQuery,
	}

	for _, q := range createTableQueries {
		if err := c.exec(q); err != nil {
			return fmt.Errorf("exec query=%s, err=%w", q, err)
		}
	}

	return nil
}

func (c *Controller) exec(query string) error {
	const createTablesTimeout = time.Second * 10

	ctx, cancel := context.WithTimeout(context.Background(), createTablesTimeout)
	defer cancel()

	_, err := c.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("exec create table query, err=%w", err)
	}

	return nil
}

func (c *Controller) Save(key []byte, data []byte) error {
	return nil
}

func (c *Controller) Update(key []byte, data []byte, revision int) error {
	return nil
}

func (c *Controller) Load(key []byte) ([]byte, int, error) {
	return nil, 0, nil
}

func (c *Controller) Delete(key []byte) error {
	return nil
}

func (c *Controller) Register(user string, password string) error {
	return nil
}

func (c *Controller) Authenticate(user string, password string) error {
	return nil
}
