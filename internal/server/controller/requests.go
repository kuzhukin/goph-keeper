package controller

const (
	createUsersTableQuery = `CREATE TABLE IF NOT EXISTS users (
		"login" text NOT NULL,
		"token" text NOT NULL,
		PRIMARY KEY ( "login" )
	);`

	createUserQuery = `INSERT INTO users (login, token) VALUES ($1, $2);`
	getUser         = `SELECT * FROM users WHERE login = $1;`
	getUserByToken  = `SELECT * FROM users WHERE token = $1;`
)

func prepareCreateUserQuery(login, password string) *query {
	return &query{request: createUserQuery, args: []any{login, password}}
}

const (
	createBinaryDataTableQuery = `CREATE TABLE IF NOT EXISTS binary_data (
		"user"			text	NOT NULL,
		"key"			text	NOT NULL,
		"value"			text	NOT NULL,
		"revision"		bigint 	NOT NULL,
		"metainfo"      text,
		PRIMARY KEY ( "user", "key" )
	);`

	// TODO: реализовать проверку наличия зарегистрированного пользователя
	addNewBinaryDataQuery      = `INSERT INTO binary_data ("user", "key", "value", "revision") VALUES ($1, $2, $3, 1);`
	updateBinaryDataQuery      = `UPDATE binary_data SET "value" = $3, revision = revision + 1 WHERE "user" = $1 AND "key" = 2;`
	getBinaryDataRevisionQuery = `SELECT "value", "revision" FROM binary_data WHERE "user" = $1 AND "key" = $2;`
	getBinaryData              = `SELECT "value", "revision" FROM binary_data WHERE "user" = $1 AND "key" = $2;`
)

type query struct {
	request string
	args    []interface{}
}

func prepareNewDataQuery(user, key, value string) *query {
	return &query{request: addNewBinaryDataQuery, args: []interface{}{user, key, value}}
}

func prepareGetDataQuery(user, key string) *query {
	return &query{request: getBinaryData, args: []interface{}{user, key}}
}
