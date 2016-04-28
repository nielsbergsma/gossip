package gossip

import (
	"time"
	"math/rand"
)

const (
	contactRate = 3
	heartbeatInterval = 1 * time.Second
	nodeTimeToLive = 5 * time.Second
)

type Agent struct {
	heartbeatPerNode map[string]*Heartbeat
	self Node
	version int
	networkClient GossipNetworkClient

	NodeAdded chan Node
	NodeRemoved chan Node
}

func NewAgent(id, uri string, capabilities []string, networkClient GossipNetworkClient) *Agent {
	return &Agent {
		heartbeatPerNode: map[string]*Heartbeat{},
		networkClient: networkClient,
		self: Node {
			Id: id,
			Uri: uri,
			Capabilities: capabilities,	
		},		
		version: 1,
		NodeAdded: make(chan Node),
		NodeRemoved: make(chan Node),
	}
}

func (a *Agent) Close() {
	close(a.NodeAdded)
	close(a.NodeRemoved)
}

func (a *Agent) Run() {
	versionIncrement := time.NewTicker(heartbeatInterval)
	receive := a.networkClient.Receiver()

	for {
		select {
		case <- versionIncrement.C:
			a.version++
			a.cleanupDeadNodes()
			a.sendHeartbeat(&Heartbeat { Node: a.self, Version: a.version })

		case message := <- receive:
			heartbeat := message.Heartbeat
			heartbeat.ReceivedAt = time.Now()

			added := a.isNewNode(heartbeat.Node)
			updated := a.processHeartbeat(heartbeat)

			if updated {
				a.sendHeartbeat(heartbeat)
			}

			if added {
				a.NodeAdded <- heartbeat.Node
			}
		}
	}

	versionIncrement.Stop()
}

func (a *Agent) Join(seeds []string) {
	send := a.networkClient.Sender()
	heartbeat := &Heartbeat { Node: a.self, Version: a.version }

	for _, seed := range seeds {
		if seed != a.self.Uri {
			send <- GossipNetworkMessage { Sender: a.self.Id, Destination: seed, Heartbeat: heartbeat }
		}
	}
}

func (a *Agent) cleanupDeadNodes() {
	timeBoundry := time.Now().Add(-nodeTimeToLive)

	for node, heartbeat := range a.heartbeatPerNode {
		if heartbeat.ReceivedAt.Before(timeBoundry) {
			delete(a.heartbeatPerNode, node)
			a.NodeRemoved <- heartbeat.Node
		}	
	}
}

func (a *Agent) processHeartbeat(heartbeat *Heartbeat) bool {
	if heartbeat.Node.Id == a.self.Id {
		return false
	}

	node, exist := a.heartbeatPerNode[heartbeat.Node.Id]
	if !exist || node.Version < heartbeat.Version {
		a.heartbeatPerNode[heartbeat.Node.Id] = heartbeat
		return true
	}

	return false
}

func (a *Agent) isNewNode(node Node) bool {
	_, exist := a.heartbeatPerNode[node.Id]
	return !exist
}

func (a *Agent) getActiveNodes() []Node {
	nodes := []Node{}

	for _, heartbeat := range a.heartbeatPerNode {
		nodes = append(nodes, heartbeat.Node)
	}
	return nodes
}

func (a *Agent) sendHeartbeat(heartbeat *Heartbeat) {
	nodes := a.getActiveNodes()

	if len(nodes) - 1 <= contactRate {
		for n := 0; n < len(nodes); n++ {
			a.sendHeartbeatTo(nodes[n].Uri, heartbeat)
		}
	} else {
		delivered := map[string]struct{}{}
		for len(delivered) < contactRate {
			index := rand.Intn(len(nodes))
			node := nodes[index]

			if _, sent := delivered[node.Id]; !sent {
				sent = a.sendHeartbeatTo(node.Uri, heartbeat)
				if sent {
					delivered[node.Id] = struct{}{}
				}
			}
		}
	}
}

func (a *Agent) sendHeartbeatTo(destination string, heartbeat *Heartbeat) bool {
	send := a.networkClient.Sender()
	if heartbeat.Node.Uri == destination {
		return false
	}

	send <- GossipNetworkMessage { Sender: a.self.Id, Destination: destination, Heartbeat: heartbeat }
	return true
}