package sqlstorage

const (
	createUserTableQuery = `CREATE TABLE IF NOT EXISTS users (
		"login" 		text NOT NULL,
		"password"		text NOT NULL,
		"crypto_key" 	text NOT NULL,
		"active"		integer NOT NULL,
		PRIMARY KEY ("login", "password")
	);`

	insertUser   = `INSERT INTO users ("login", "password", "crypto_key", "active") VALUES ($1, $2, $3, 1);`
	getUser      = `SELECT "login", "password", "crypto_key" FROM users WHERE "active" == 1;`
	changeActive = `UPDATE users SET "active" = 0 WHERE "user" = $1`
)

func prepareInsertUserQuery(login, password, crypto_key string) *query {
	return &query{request: insertUser, args: []interface{}{login, password, crypto_key}}
}

func prepareGetUserQuery() *query {
	return &query{request: getUser}
}

func prepareChangeActiveQuery(login string) *query {
	return &query{request: changeActive, args: []any{login}}
}
