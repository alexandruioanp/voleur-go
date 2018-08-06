package main

import (
	"net/http"
	"log"
)

func web_listen(events_in chan []byte, web_update_out chan VoleurUpdateType) {
	broker := NewSSEServer(events_in)
	api_handler := NewAPIHandler(web_update_out)

	http.Handle("/events", broker)
	http.Handle("/", http.FileServer(http.Dir("./js")))
	http.Handle("/volOps", api_handler)
	
//	different path?
//	log.Fatal(http.ListenAndServe(":8080", http.FileServer(http.Dir("../src/voleur/js"))))
	go func() { log.Fatal(http.ListenAndServe(":8080", nil)) } ()
	
}
