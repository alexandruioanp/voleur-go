package main

import (
	"fmt"
	"time"
	"voleur/ifaces"
//	"os"
//	"log"
)

func main() {
	fmt.Println("main")

//	dir, err := os.Getwd()
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(dir)
	
	change_chan := make(chan []byte)
	web_req_ch := make(chan ifaces.VoleurUpdateType)

	var audio_interface ifaces.IControlInterface = ifaces.NewPulseCMDLineInteface()
	go audio_interface.Listen(change_chan)
	go audio_interface.ApplyChanges(web_req_ch)

	go web_listen(change_chan, web_req_ch, audio_interface)

	for {
		time.Sleep(time.Second)
	}

	fmt.Println("main done")
}
