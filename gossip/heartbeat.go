package gossip

import (
	"time"
)

type Heartbeat struct {
	Node Node `json:"node"`
	Version int `json:"version"`
	ReceivedAt time.Time `json:"-"`
}
