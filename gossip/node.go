package gossip

type Node struct {
	Id string `json:"id"`
	Uri string `json:"uri"`
	Capabilities []string `json:"capacities"`
}