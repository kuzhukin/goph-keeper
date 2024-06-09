package endpoint

// endpoints
const (
	// POST - registrer new user
	RegisterEndpoint = "/api/user/register"
	// POST - auth user
	AuthEndpoint = "/api/user/auth"
	// PUT - load new data to storage
	// POST - update binary data to storage
	// GET - get binary data from storage
	// DELETE - delete item from storage
	BinaryDataEndpoint = "/api/data/binary"
	// PUT, POST, GET, DELETE
	// Key = Value secret
	SecretEndpoint = "/api/data/secret"
	// PUT, POST, GET, DELETE
	// card data
	WalletEndpoint = "/api/data/wallet"
)
