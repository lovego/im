package im

type message struct {
	System   string
	Users    []string
	Business string
	Version  string
}

func (im *IM) Push(system string, users []string, business string) error {
	// 1. record version
	// 2. publish message
	return nil
}
