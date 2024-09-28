package data

type Getter[K, T any] interface {
	Get(key K) (T, bool, error)
	GetAll() ([]T, error)
}

type Adder[K, T any] interface {
	Add(item T) error
	AddAll(items []T) error
}

type Deleter[K, T any] interface {
	Delete(key K)
	DeleteAll()
}

type Repository[K, T any] interface {
	Getter[K, T]
	Adder[K, T]
	Deleter[K, T]
}
