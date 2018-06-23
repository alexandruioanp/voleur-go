package main

import (  
    "fmt"
    "os/exec"
	"bufio"
	"os"
    "encoding/json"
	"regexp"
	"errors"
    "strings"
    "strconv"
)

//const PA_VOLUME_MAX = 65536

type UpdateType int

const (  // iota is reset to 0
        Update UpdateType = iota  // c0 == 0
        Add
        Remove
)

type VoleurUpdate struct {
    Name      string
    Vol       int
    IsSinkVol bool
    Type	  UpdateType 
}

func listen(change_event_out chan string) {
	cmd := exec.Command("pactl", "subscribe", "change")
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		os.Exit(1)
	}
	
	scanner := bufio.NewScanner(cmdReader)
	
	go func() {
		for scanner.Scan() {
			change_event_out <- scanner.Text()
//			fmt.Printf("%s\n", scanner.Text())
		}
	}()
	
	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		os.Exit(1)
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
		os.Exit(1)
	}
}

func pactl_get_sinkinput_details(sinkinput_num string) (VoleurUpdate, error) {
	cmd_out, err := exec.Command("pactl", "list", "sink-inputs").Output()
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
	
	out := VoleurUpdate{Name: "", Vol: 0, IsSinkVol: false}
	
	regex_app_name, err := regexp.Compile(`application.name = "(.*)"`)
	if err != nil {
		return out, errors.New("Regex error")
	}
	
	regex_volume, err := regexp.Compile(`front-left: .+ (\d+)%`)
	if err != nil {
		return out, errors.New("Regex error")
	}
	
	//TODO use icon-name
	
	for _, el := range strings.Split(string(cmd_out), "Sink Input #")[1:] {
		first_line := strings.Split(el, "\n")[0]
		if first_line == sinkinput_num {
			app_name := regex_app_name.FindStringSubmatch(el)[1]
			vol_left := regex_volume.FindStringSubmatch(el)[1]
//			fmt.Println(app_name)
//			fmt.Println(vol_left)
			out.Name = app_name
			out.Vol, _ = strconv.Atoi(vol_left) 
			out.Type = Update
			// found sinkinput #sinkinput_num
			//TODO parse appname, volume
		}
		
	}
	
	return out, nil
}

func parse_pactl_update_msg(str string) (VoleurUpdate, error) {	
	var out VoleurUpdate
//	
	if strings.Contains(str, "'change' on sink-input") {
		r, err := regexp.Compile(`[\d]+`)
		
//		fmt.Println(str)
		
		if err != nil {
			return out, errors.New("Regex error")
		}
		
		// TODO check if sinkinput or sink update
		sinkinput_num := r.FindString(str)
		
		out, err = pactl_get_sinkinput_details(sinkinput_num)
		if err != nil {
			return out, err
		}
	} else {
		return out, errors.New("Not the update you're looking for")
	}
	
	return out, nil
}

func decode(change_in chan string, json_out chan []byte) {
	for {
		str := <- change_in
		update_msg, err := parse_pactl_update_msg(str)
		if err != nil {
			continue
		}
		b, err := json.Marshal(update_msg)
		if err != nil {
			continue
		}
		json_out <- b
	}
}

func main() {
	change_chan := make(chan string)
    go listen(change_chan)
    dec_chan := make(chan []byte)
    go decode(change_chan, dec_chan)

	var m VoleurUpdate
    
    for
    {
		err := json.Unmarshal(<- dec_chan, &m)
		if err == nil {
	    	fmt.Println(m)
		}
	    fmt.Println("main function")
    }
}