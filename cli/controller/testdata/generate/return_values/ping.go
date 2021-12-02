package ping

type Ping struct{}

func New() (*Ping, error) {
	return Ping{}
}
