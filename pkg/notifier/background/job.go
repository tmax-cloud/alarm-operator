package background

type Job interface {
	Execute(job interface{}) error
}
