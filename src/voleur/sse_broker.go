package main

import (
	"fmt"
	"log"
	"net/http"
	"voleur/ifaces"
)

// https://robots.thoughtbot.com/writing-a-server-sent-events-server-in-go
type Broker struct {
	// Events are pushed to this channel by the main events-gathering routine
	Notifier chan []byte

	// New client connections
	newClients chan chan []byte

	// Closed client connections
	closingClients chan chan []byte

	// Client connections registry
	clients map[chan []byte]bool

	audio_interface ifaces.IControlInterface
}

// Listen on different channels and act accordingly
func (broker *Broker) listen() {
	for {
		select {
		case s := <-broker.newClients:
			// A new client has connected.
			// Register their message channel
			// TODO: send sink & sinkinput list
			broker.clients[s] = true
			log.Printf("Client added. %d registered clients", len(broker.clients))
			toggles := broker.audio_interface.GetAll()

			for _, toggle := range toggles {
				s <- toggle
			}

		case s := <-broker.closingClients:
			// A client has dettached and we want to
			// stop sending them messages.
			delete(broker.clients, s)
			log.Printf("Removed client. %d registered clients", len(broker.clients))

		case event := <-broker.Notifier:
			//			log.Println("Broadcasting")
			// We got a new event from the outside!
			// Send event to all connected clients
			for clientMessageChan, _ := range broker.clients {
				clientMessageChan <- event
			}
		}
	}
}

func (broker *Broker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	flusher, ok := rw.(http.Flusher)

	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Set the headers related to event streaming.
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the Broker's connections registry
	messageChan := make(chan []byte)

	// Signal the broker that we have a new connection
	broker.newClients <- messageChan

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		broker.closingClients <- messageChan
	}()

	// Listen to connection close and un-register messageChan
	notify := rw.(http.CloseNotifier).CloseNotify()

	go func() {
		<-notify
		broker.closingClients <- messageChan
	}()

	// block waiting for messages broadcast on this connection's messageChan
	for {
		// Write to the ResponseWriter
		// Server Sent Events compatible
		data := <-messageChan
//		fmt.Printf("sending `data: %s`\n", data)
		//		fmt.Fprintln(rw, data)
		fmt.Fprintf(rw, "data: %s\n\n", data)

		// Flush the data immediatly instead of buffering it for later.
		flusher.Flush()
	}
}

// Broker factory
func NewSSEServer(events_in_chan chan []byte, audio_interface ifaces.IControlInterface) (broker *Broker) {
	// Instantiate a broker
	broker = &Broker{
		Notifier:        events_in_chan,
		newClients:      make(chan chan []byte),
		closingClients:  make(chan chan []byte),
		clients:         make(map[chan []byte]bool),
		audio_interface: audio_interface,
	}

	//	fmt.Printf("created broker")
	// Set it running - listening and broadcasting events
	go broker.listen()

	return
}
