package sse

type Event struct {
	ClientID string
	Data     interface{}
}

func (e Event) IsBroadCast() bool {
	return e.ClientID == ""
}
