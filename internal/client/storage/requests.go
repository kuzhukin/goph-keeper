package storage

type query struct {
	request string
	args    []interface{}
}

const (
	createUserTableQuery = `CREATE TABLE IF NOT EXISTS users (
		"login" 	text NOT NULL,
		"password"	text NOT NULL,
		"active"	integer NOT NULL,
		PRIMARY KEY ("login", "password")
	);`

	insertUser = `INSERT INTO users ("login", "password", "active") VALUES ($1, $2, 1);`
	getUser    = `SELECT "login", "password" FROM users WHERE "active" == 1;`
)

func prepareInsertUserQuery(login, password string) *query {
	return &query{request: insertUser, args: []interface{}{login, password}}
}

func prepareGetUserQuery() *query {
	return &query{request: getUser}
}

const (
	createDataTableQuery = `CREATE TABLE IF NOT EXISTS data (
		"key"			text	NOT NULL,
		"value"			text	NOT NULL,
		"revision"		bigint 	NOT NULL,
		PRIMARY KEY ( "key" )
	);`

	// TODO: реализовать проверку наличия зарегистрированного пользователя
	addNewDataQuery  = `INSERT INTO data ("key", "value", "revision") VALUES ($1, $2, 1);`
	updateDataQuery  = `UPDATE data SET "value" = $3, revision = revision + 1 WHERE "user" = $1 AND "key" = 2;`
	getRevisionQuery = `SELECT "value", "revision" FROM data WHERE "key" = $2;`
	getData          = `SELECT "value", "revision" FROM data WHERE "key" = $2;`
)
