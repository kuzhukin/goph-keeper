package controller

const (
	createUsersTableQuery = `CREATE TABLE IF NOT EXISTS users (
		"login"		text NOT NULL,
		"password"	text NOT NULL,
		PRIMARY KEY ( "login" )
	);`

	createUserQuery = `INSERT INTO users (login, password) VALUES ($1, $2);`
	getUser         = `SELECT "login", "password" FROM users WHERE login = $1;`
	getUserByToken  = `SELECT * FROM users WHERE password = $1;`
)

func prepareCreateUserQuery(login, password string) *query {
	return &query{request: createUserQuery, args: []any{login, password}}
}

func prepareGetUserQuery(login string) *query {
	return &query{request: getUser, args: []any{login}}
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

	addNewBinaryDataQuery = `INSERT INTO binary_data ("user", "key", "value", "revision") VALUES ($1, $2, $3, 1);`
	updateBinaryDataQuery = `UPDATE binary_data SET "value" = $3, revision = $4 WHERE "user" = $1 AND "key" = $2;`
	getBinaryData         = `SELECT "value", "revision" FROM binary_data WHERE "user" = $1 AND "key" = $2;`
	deleteBinaryData      = `DELETE FROM binary_data WHERE "user" = $1 AND "key" = $2;`
)

type query struct {
	request string
	args    []any
}

func prepareNewDataQuery(user, key, value string) *query {
	return &query{request: addNewBinaryDataQuery, args: []any{user, key, value}}
}

func prepareGetDataQuery(user, key string) *query {
	return &query{request: getBinaryData, args: []any{user, key}}
}

func prepareUpdateDataQuery(user, key, data string, revision uint64) *query {
	return &query{request: updateBinaryDataQuery, args: []any{user, key, data, revision}}
}

func prepareDeleteDataQuery(user, key string) *query {
	return &query{request: deleteBinaryData, args: []any{user, key}}
}
