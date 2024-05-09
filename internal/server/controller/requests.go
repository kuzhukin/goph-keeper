package controller

const (
	createUsersTableQuery = `CREATE TABLE IF NOT EXIST users (
		login text NOT NULL,
		token text NOT NULL,
		PRIMARY KEY ( login )
	);`

	createUserQuery = `INSERT INTO users (login, token) VALUES ($1, $2);`
	getUser         = `SELECT * FROM users WHERE login = $1;`
	getUserByToken  = `SELECT * FROM users WHERE token = $1;`
)

const (
	createDataTableQuery = `CREATE TABLE IF NOT EXIST data (
		user_login text NOT NULL,
		key text NOT NULL,
		value text NOT NULL,
		revision int NOT NULL,
		PRIMARY KEY ( user, key )
	);`

	addNewDataQuery  = `INSERT INTO data (user, key, value, revision) VALUES ($1, $2, $3, 1);`
	updateDataQuery  = `UPDATE data SET value = $3, revision = revision + 1 WHERE user = $1 AND key = 2;`
	getRevisionQuery = `SELECT revision FROM data WHERE user = $1 AND key = $2;`
	getData          = `SELECT value, revision FROM data WHERE user = $1 AND key = $2;`
)
