package gossip

import (
	"net"
	"net/url"
	"encoding/json"
)

type UDPGossipNetworkClient struct {
	protocol string
	bindUri string
	sender chan GossipNetworkMessage
	receiver chan GossipNetworkMessage
}

func NewUDPGossipNetworkClient(bindUri string) *UDPGossipNetworkClient {
	return &UDPGossipNetworkClient {
		protocol: "udp",
		bindUri: bindUri,
		sender: make(chan GossipNetworkMessage),
		receiver: make(chan GossipNetworkMessage),
	}
}

func (gnc *UDPGossipNetworkClient) Sender() chan GossipNetworkMessage {
	return gnc.sender
}

func (gnc *UDPGossipNetworkClient) Receiver() chan GossipNetworkMessage {
	return gnc.receiver
}

func (gnc *UDPGossipNetworkClient) Close() {
	close(gnc.sender)
	close(gnc.receiver)
}

func (gnc *UDPGossipNetworkClient) Run() {
	go gnc.receive()

	for message := range gnc.sender {
		gnc.send(&message)
	}
}

func (gnc *UDPGossipNetworkClient) receive() error {
	bindUrl, err := url.Parse(gnc.bindUri)
	if err != nil {
		return err
	}

	bindAddress,err := net.ResolveUDPAddr(gnc.protocol, bindUrl.Host)
    if err != nil {
		return err	
	}
 
    connection, err := net.ListenUDP(gnc.protocol, bindAddress)
    if err != nil {
		return err	
	}
    defer connection.Close()

    buffer := make([]byte, 1024)
    for {
    	length, _, err := connection.ReadFromUDP(buffer)
    	if err != nil {
    		continue
		}

		message := GossipNetworkMessage{}
		err = json.Unmarshal(buffer[0:length], &message)
		if err != nil {
    		continue
		}

		gnc.receiver <- message
    }

    return nil
}

func (gnc *UDPGossipNetworkClient) send(message *GossipNetworkMessage) error {
	senderUrl, err := url.Parse(message.Sender)
	if err != nil {
		return err
	}

	senderAddress, err := net.ResolveUDPAddr(gnc.protocol, senderUrl.Host)
	if err != nil {
		return err
	}

	destinationUrl, err := url.Parse(message.Destination)
	if err != nil {
		return err
	}

	destinationAddress, err := net.ResolveUDPAddr(gnc.protocol, destinationUrl.Host)
	if err != nil {
		return err
	}

	connection, err := net.DialUDP(gnc.protocol, senderAddress, destinationAddress)
	if err != nil {
		return err
	}
	defer connection.Close()

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = connection.Write(data)
	if err != nil {
		return err
	}

	return nil
}