package worker

import "time"

type Event struct {
	Type      string
	Kind      string
	Name      string
	Namespace string
	Timestamp time.Time
	Message   string
	Reason    string
	Source    string
	Host      string
	count     int32
}
