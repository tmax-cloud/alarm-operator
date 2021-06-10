package notification

type Registry interface {
	Save(id string, data []byte) error
	Load(id string) ([]byte, error)
}

type Queue interface {
	Push(data []byte) error
	Pop() ([]byte, error)
}
