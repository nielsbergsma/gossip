package gossip

type GossipNetworkMessage struct {
	Sender string
	Destination string
	Heartbeat *Heartbeat
}

type GossipNetworkClient interface {
	Sender() chan GossipNetworkMessage
	Receiver() chan GossipNetworkMessage
	Close()
}