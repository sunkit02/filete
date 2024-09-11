package data

import "time"

type Message struct {
	Id       uint64    `json:"id"`
	Title    string    `json:"title"`
	Body     string    `json:"body"`
	TimeSent time.Time `json:"timeSent"`
}

type Identity struct {
	Secret string `json:"secret"`
	Alias  string `json:"alias"`
}

type LogString interface {
	LogStr() string
}
