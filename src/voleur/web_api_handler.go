package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"voleur/ifaces"
)

type APIHandler struct {
	// POST data is processed and pushed to this channel
	VolumeEventChannel chan ifaces.VoleurUpdateType
}

func (api_handler *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//	fmt.Printf("URL path: %s\n", r.URL.Path)

	switch r.Method {

	// TODO
	case "GET":
		fmt.Print("GET ")
		fmt.Printf("%v %v\n", r.URL, r.URL.Path)//

	case "POST":
		w.WriteHeader(http.StatusOK)

		upd := ifaces.VoleurUpdateType{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&upd)

		if err == nil {
			api_handler.VolumeEventChannel <- upd
		}
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

// API factory
func NewAPIHandler(event_channel chan ifaces.VoleurUpdateType) (api_handler *APIHandler) {
	// Instantiate a handler
	api_handler = &APIHandler{
		VolumeEventChannel: event_channel,
	}

	return
}
