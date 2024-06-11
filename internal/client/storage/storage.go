package storage

type Storage interface {
	UserStorage
	DataStorage
	WalletStorage
	SecretStorage
	Stop() error
}

type DataStorage interface {
	CreateData(u *User, r *Record) error
	UpdateData(u *User, r *Record) (uint64, bool, error)
	LoadData(u *User, name string) (*Record, error)
	ListData(u *User) ([]*Record, error)
	DeleteData(u *User, r *Record) error
}

type UserStorage interface {
	Register(login string, password string, token string, cryptokey string) error
	GetActive() (*User, error)
}

type WalletStorage interface {
	CreateCard(u *User, c *BankCard) (string, error)
	ListCard(u *User) ([]*BankCard, error)
	DeleteCard(u *User, cardNumber string) error
}

type SecretStorage interface {
	CreateSecret(u *User, s *Secret) (string, error)
	GetSecret(u *User, secretKey string) (*Secret, error)
	DeleteSecret(u *User, secretKey string) error
}
