package data

import "time"

type Message struct {
	Id       MessageId `json:"id"`
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

type MessageId uint64
