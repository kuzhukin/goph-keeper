package storage

type query struct {
	request string
	args    []interface{}
}

const (
	createUserTableQuery = `CREATE TABLE IF NOT EXISTS users (
		"login" 		text NOT NULL,
		"password"		text NOT NULL,
		"crypto_key" 	text NOT NULL,
		"active"		integer NOT NULL,
		PRIMARY KEY ("login", "password")
	);`

	insertUser = `INSERT INTO users ("login", "password", "crypto_key", "active") VALUES ($1, $2, $3, 1);`
	getUser    = `SELECT "login", "password", "crypto_key" FROM users WHERE "active" == 1;`
)

func prepareInsertUserQuery(login, password, crypto_key string) *query {
	return &query{request: insertUser, args: []interface{}{login, password, crypto_key}}
}

func prepareGetUserQuery() *query {
	return &query{request: getUser}
}

const (
	createDataTableQuery = `CREATE TABLE IF NOT EXISTS data (
		"user"			text		NOT_NULL,
		"key"			text		NOT NULL,
		"value"			text		NOT NULL,
		"revision"		integer 	NOT NULL,
		PRIMARY KEY ( "user", "key" )
	);`

	addNewDataQuery  = `INSERT INTO data ("user", "key", "value", "revision") VALUES ($1, $2, $3, 1);`
	updateDataQuery  = `UPDATE data SET "value" = $3 WHERE "user" = $1 AND "key" = 2;`
	getRevisionQuery = `SELECT "revision" FROM data WHERE "user" = $1 AND "key" = $2;`
	getData          = `SELECT "value", "revision" FROM data WHERE "user" = $1 AND "key" = $2;`
	listData         = `SELECT "key", "value", "revision" FROM data WHERE "user" = $1;`
)

func prepareAddDataQuery(user string, key string, value string) *query {
	return &query{request: addNewDataQuery, args: []any{user, key, value}}
}

func prepareUpdateDataQuery(user, key, value string) *query {
	return &query{request: updateDataQuery, args: []any{user, key, value}}
}

func preareGetRevisionQuery(user, key string) *query {
	return &query{request: getRevisionQuery, args: []any{user, key}}
}

func prepareGetDataQuery(user, key string) *query {
	return &query{request: getData, args: []any{user, key}}
}

func prepareListDataQuery(user string) *query {
	return &query{request: listData, args: []any{user}}
}
