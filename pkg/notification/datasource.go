package notification

type Registry interface {
	Save(id string, namespace string, data []byte) error
	Load(id string, namespace string) ([]byte, error)
}

type Queue interface {
	Push(data []byte) error
	Pop() ([]byte, error)
}
