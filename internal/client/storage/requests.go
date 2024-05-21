package storage

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
