package port

type MetaCacheController interface {
	Add(key string, value []byte)
	Get(key string) ([]byte, bool)
	Delete(key string)
}
