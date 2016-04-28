package main

import (
	"log"
	"fmt"
	"flag"
	"time"

	"./gossip"
)

func main() {
	bindUri := flag.String("bind", "", "Bind uri (e.g.: udp://127.0.0.0.1:3000)")
	flag.Parse()

	if *bindUri == "" {
		log.Fatalf("Error: no bind uri supplied")
		return
	}

	log.Printf("Starting agent at %v", *bindUri)

	id := fmt.Sprintf("agent#%v", time.Now().Unix())
	networkClient := gossip.NewUDPGossipNetworkClient(*bindUri)
	go networkClient.Run()

	agent := gossip.NewAgent(id, *bindUri, []string{ "1", "3" }, networkClient)
	go agent.Run()

	//join existing cluster
	agent.Join([]string { "udp://127.0.0.1:3001" })

	for {
		select {
		case node := <- agent.NodeAdded: 
			log.Printf("Node added %+v", node)

		case node := <- agent.NodeRemoved:
			log.Printf("Node removed %+v", node)

		}
	}	
}
