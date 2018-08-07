package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type APIHandler struct {
	// POST data is processed and pushed to this channel
	VolumeEventChannel chan VoleurUpdateType
}

func post_to_volupdate(data url.Values) (v VoleurUpdateType) {
	name := data.Get("name")
	vol, err := strconv.Atoi(data.Get("value"))
	if err != nil {
		return
	}

	isSinkVol, err := strconv.ParseBool(data.Get("isSinkVol"))
	if err != nil {
		return
	}

	v.Name = name
	v.Vol = vol
	v.IsSinkVol = isSinkVol
	v.Type = Update

	return
}

func (api_handler *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("URL path: %s\n", r.URL.Path)

	switch r.Method {

	case "GET":
		fmt.Println("GET")
		fmt.Println(r.URL.Query())

	case "POST":
		fmt.Println("POST")
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		//		fmt.Println(r.Form)
		w.WriteHeader(http.StatusOK)
		api_handler.VolumeEventChannel <- post_to_volupdate(r.PostForm)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

// API factory
func NewAPIHandler(event_channel chan VoleurUpdateType) (api_handler *APIHandler) {
	// Instantiate a handler
	api_handler = &APIHandler{
		VolumeEventChannel: event_channel,
	}

	return
}
