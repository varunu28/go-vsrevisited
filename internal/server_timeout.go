package internal

import "time"

type ServerTimeout struct {
	Timeout         *time.Ticker
	Reset           chan struct{}
	TimeoutInterval int
}

// NewServerTimeout creates a new instance of ServerTimeout struct
func NewServerTimeout(timeoutInterval int) *ServerTimeout {
	return &ServerTimeout{
		Timeout:         time.NewTicker(time.Duration(timeoutInterval) * time.Millisecond),
		Reset:           make(chan struct{}),
		TimeoutInterval: timeoutInterval,
	}
}
