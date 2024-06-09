package storage

type Storage interface {
	UserStorage
	DataStorage
	WalletStorage
	Stop() error
}

type DataStorage interface {
	SaveData(u *User, r *Record) (uint64, error)
	LoadData(u *User, name string) (*Record, error)
	ListData(u *User) ([]*Record, error)
	DeleteData(u *User, r *Record) error
}

type UserStorage interface {
	Register(login string, password string, cryptokey string) error
	GetActive() (*User, error)
}

type WalletStorage interface {
	CreateCard(u *User, c *BankCard) error
	ListCard(u *User) ([]*BankCard, error)
	DeleteCard(u *User, cardNumber string) error
}
