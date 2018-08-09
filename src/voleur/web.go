package main

import (
	"log"
	"net/http"
	"regexp"
	"voleur/ifaces"
)

var path_regexp = regexp.MustCompile(`^\/volOps.*`)
var api_handler *APIHandler
var file_server http.Handler

func myRouter(w http.ResponseWriter, r *http.Request) {
	regexp_res := path_regexp.FindString(r.RequestURI)
	if regexp_res == "" {
		file_server.ServeHTTP(w, r)
	} else {
		api_handler.ServeHTTP(w, r)
	}
}

func web_listen(events_in chan []byte,
				web_update_out chan ifaces.VoleurUpdateType,
				audio_interface ifaces.IAudioInterface) {
	broker := NewSSEServer(events_in, audio_interface)
	api_handler = NewAPIHandler(web_update_out)
//	file_server = http.FileServer(http.Dir("./js")) //	different working dir
	file_server = http.FileServer(http.Dir("./src/voleur/js"))

	http.Handle("/events", broker)
	http.HandleFunc("/", myRouter)

	go func() { log.Fatal(http.ListenAndServe(":8080", nil)) }()

}
