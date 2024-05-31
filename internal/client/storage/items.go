package storage

type Record struct {
	Name     string
	Data     string
	Revision uint64
}

type User struct {
	Login    string
	Password string
	IsActive bool
}
