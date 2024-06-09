package sqlstorage

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
	deleteBinaryData = `DELETE FROM data WHERE "user" = $1 AND "key" = $2;`
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

func prepareDeleteDataQuery(user string, dataKey string) *query {
	return &query{request: deleteBinaryData, args: []any{user, dataKey}}
}

const (
	createCardTableQuery = `CREATE TABLE IF NOT EXISTS cards (
		"user"			text		NOT_NULL,
		"number			text		NOT NULL,
		"data"			text		NOT NULL,
		PRIMARY KEY ( "user", "number" )
	);`

	addCard    = `INSERT INTO cards ("user", "number", "data") VALUES ($1, $2, $3);`
	listCard   = `SELECT "data" FROM cards WHERE "user" = $1;`
	deleteCard = `DELETE FROM cards WHERE "user" = $1 AND "number" = $2;`
)

func prepareAddCard(user, number, data string) *query {
	return &query{request: addCard, args: []any{user, number, data}}
}

func prepareListCard(user string) *query {
	return &query{request: listCard, args: []any{user}}
}

func prepareDeleteCard(user, number string) *query {
	return &query{request: deleteCard, args: []any{user, number}}
}
