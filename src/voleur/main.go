package main

import (
	"fmt"
	"voleur/ifaces"
	"time"
)

func main() {
	fmt.Println("main")
	change_chan := make(chan []byte)
	web_req_ch := make(chan ifaces.VoleurUpdateType)

	go web_listen(change_chan, web_req_ch)

	var audio_interface ifaces.IAudioInterface = ifaces.NewPulseCMDLineInteface()
	go audio_interface.Listen(change_chan)
	go audio_interface.ApplyChanges(web_req_ch)

	for {
        time.Sleep(time.Second)
	}

	fmt.Println("main done")
}
