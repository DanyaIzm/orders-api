package storage

type Storage interface {
	Close()
	OrderRepo() OrderRepo
}
