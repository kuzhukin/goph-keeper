package controller

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kuzhukin/goph-keeper/internal/server/handler"
	"github.com/kuzhukin/goph-keeper/internal/zlog"
)

var _ handler.Storage = &Controller{}
var _ handler.Authenticator = &Controller{}
var _ handler.Registrator = &Controller{}

var (
	ErrDataAlreadyExist = errors.New("data already exist")
)

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

func (c *Controller) Save(ctx context.Context, user, key, data string) error {
	executor := c.makeExecFunc(ctx, prepareNewDataQuery(user, key, data), time.Second*5)
	_, err := doQuery(executor)
	if err != nil {
		if isNotUniqueError(err) {
			return ErrDataAlreadyExist
		}

		return fmt.Errorf("do query err=%w", err)
	}

	return nil
}

func (c *Controller) Update(ctx context.Context, user, key, data string, revision uint64) error {
	return nil
}

func (c *Controller) Load(ctx context.Context, user, key string) (string, uint64, error) {
	executor := c.makeQueryFunc(ctx, prepareGetDataQuery(user, key), time.Second*5)
	rows, err := doQuery(executor)
	if err != nil {
		return "", 0, fmt.Errorf("do query, err=%w", err)
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			zlog.Logger().Errorf("rows close err=%s", err)
		}
	}()

	var data string
	var revision uint64

	err = scanRow(rows, &data, &revision)
	if err != nil {
		return "", 0, fmt.Errorf("user=%s, key=%s, err=%w", user, key, err)
	}

	return data, revision, nil
}

func (c *Controller) Delete(ctx context.Context, user, key string) error {
	return nil
}

func (c *Controller) Register(ctx context.Context, user string, password string) error {
	return nil
}

func (c *Controller) Authenticate(ctx context.Context, user string, password string) error {
	return nil
}

// ----------------------------------------------------------------------------------------------
// -------------------------------------- Internal Methods --------------------------------------
// ----------------------------------------------------------------------------------------------

func scanRow(rows *sql.Rows, dest ...any) error {
	if rows.Next() {
		err := rows.Scan(dest...)
		if err != nil {
			return fmt.Errorf("scan rows err=%w", err)
		}
	} else {
		return fmt.Errorf("empty rows")
	}

	return rows.Err()
}

func (c *Controller) makeExecFunc(ctx context.Context, query *query, timeout time.Duration) func() (*sql.Result, error) {
	return func() (r *sql.Result, err error) {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		res, err := c.db.ExecContext(ctxWithTimeout, query.request, query.args...)
		if err != nil {
			return nil, fmt.Errorf("exec query=%v err=%w", query, err)
		}

		return &res, nil
	}
}

func (c *Controller) makeQueryFunc(ctx context.Context, query *query, timeout time.Duration) func() (*sql.Rows, error) {
	return func() (*sql.Rows, error) {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		rows, err := c.db.QueryContext(ctxWithTimeout, query.request, query.args...)
		if err != nil {
			return nil, fmt.Errorf("query metric, err=%w", err)
		}

		return rows, nil
	}
}

var tryingIntervals = []time.Duration{
	time.Millisecond * 100,
	time.Millisecond * 300,
	time.Millisecond * 500,
}

func doQuery[T any](queryFunc func() (*T, error)) (*T, error) {
	var commonErr error
	max := len(tryingIntervals)

	for trying := 0; trying <= max; trying++ {
		rows, err := queryFunc()
		if err != nil {
			commonErr = errors.Join(commonErr, err)

			if trying < max && isRetriableError(err) {
				time.Sleep(tryingIntervals[trying])
				continue
			}

			return nil, commonErr
		}

		return rows, nil
	}

	return nil, commonErr
}

func isRetriableError(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)
}

func isNotUniqueError(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation
}

func doTransactionQuery[T any](ctx context.Context, tx *sql.Tx, query *query, parseFun func(rows *sql.Rows) (T, error)) (T, error) {
	var result T

	rows, err := tx.QueryContext(ctx, query.request, query.args...)
	if err != nil {
		return result, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			zlog.Logger().Errorf("rows close err=%s", err)
		}
	}()

	result, err = parseFun(rows)
	if err != nil {
		return result, err
	}

	if err := rows.Err(); err != nil {
		return result, err
	}

	return result, nil
}
