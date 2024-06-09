package sql

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
