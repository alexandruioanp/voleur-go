package ifaces

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type sinkInputInfo struct {
	Name      string
	Vol       int
	SI_number string
}

type PulseCMDLineInterface struct {
	sink_input_cache map[string]sinkInputInfo
}

const PA_VOLUME_MAX = 65536

func (pulse_iface *PulseCMDLineInterface) Listen(json_out chan []byte) {
	change_chan := make(chan string)
	go pulse_iface.listen_pulse(change_chan)
	go pulse_iface.decode(change_chan, json_out)
}

func (pulse_iface *PulseCMDLineInterface) listen_pulse(change_event_out chan string) {
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

func (pulse_iface *PulseCMDLineInterface) get_cached_sinkinput_details(si_num string) (sinkInputInfo, bool) {
	si_info, ok := pulse_iface.sink_input_cache[si_num]
	return si_info, ok
}

func pactl_get_sinkinput_volume(sinkinput_num string) (int, error) {
	cmd_out, err := exec.Command("pactl", "list", "sink-inputs").Output()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		return -1, err
	}

	//TODO use icon-name ?
	regex_volume := regexp.MustCompile(`front-left: .+ (\d+)%`)

	for _, el := range strings.Split(string(cmd_out), "Sink Input #")[1:] {
		first_line := strings.Split(el, "\n")[0]
		if first_line == sinkinput_num {
			vol_left := regex_volume.FindStringSubmatch(el)[1]
			vol, err := strconv.Atoi(vol_left)
			return vol, err
		}
	}

	return -1, errors.New("Cannot get volume")
}

func si_details_to_update(si_info sinkInputInfo) (upd VoleurUpdateType) {
	upd.Name = si_info.Name
	if upd.AuxData == nil {
		upd.AuxData = make(map[string]string)
	}
	upd.AuxData["SI_number"] = si_info.SI_number

	return upd
}

func (pulse_iface *PulseCMDLineInterface) parse_event(str string) (VoleurUpdateType, error) {
	var out VoleurUpdateType
	
	// TODO handle new, remove sink-input
	// TODO handle sink volume
	if strings.Contains(str, "'change' on sink-input") {
		r := regexp.MustCompile(`[\d]+`)

		// TODO check if sinkinput or sink update
		sinkinput_num := r.FindString(str)
		cached_details, ok := pulse_iface.sink_input_cache[sinkinput_num]

		out = si_details_to_update(cached_details)
		out.Type = Update
		out.AuxData["isSinkVol"] = "0"

		if !ok {
			return out, errors.New("Failed to get sinkinput details")
		}

		vol, err := pactl_get_sinkinput_volume(sinkinput_num)
		out.Vol = vol

		if err == nil {
			return out, nil
		} else {
			return VoleurUpdateType{}, err
		}
	} else if strings.Contains(str, "'new' on sink-input") {
		// TODO optimise?
		fmt.Println("new si")
		pulse_iface.build_sink_input_cache()
		// TODO send update to clients
		return out, errors.New("Not implemented yet") 
	} else if strings.Contains(str, "'remove' on sink-input") {
		// TODO optimise?
		fmt.Println("si removed")
		pulse_iface.build_sink_input_cache()
		// TODO send update to clients
		return out, errors.New("Not implemented yet")
	} else {
		return out, errors.New("Not the update you're looking for")
	}

	return out, nil
}

func (pulse_iface *PulseCMDLineInterface) decode(change_in chan string, json_out chan []byte) {
	for {
		str := <-change_in
		update_msg, err := pulse_iface.parse_event(str)
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

func (pulse_iface *PulseCMDLineInterface) ApplyChanges(in_chan chan VoleurUpdateType) {
	for {
		update := <-in_chan
		vol_to_set := (PA_VOLUME_MAX * update.Vol) / 100
		fmt.Println(update)
		si_number, ok := update.AuxData["SI_number"]
		if !ok {
			fmt.Println("Missing SI_number")
			continue
		}
		
		if _, ok := pulse_iface.sink_input_cache[si_number]; !ok {
			continue
		}
		cmd := exec.Command("pactl", "set-sink-input-volume", si_number, strconv.Itoa(vol_to_set))

		var stderr bytes.Buffer
		//	cmd.Stdout = &out
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error running Cmd", stderr.String())
		}
	}
}

func (pulse_iface *PulseCMDLineInterface) build_sink_input_cache() {
	pulse_iface.sink_input_cache = make(map[string]sinkInputInfo)

	cmd_out, err := exec.Command("pactl", "list", "sink-inputs").Output()
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}

	for _, el := range strings.Split(string(cmd_out), "Sink Input #")[1:] {
		si_info := sinkInputInfo{}

		first_line := strings.Split(el, "\n")[0]
		si_num_r := regexp.MustCompile(`[\d]+`)

		regex_app_name := regexp.MustCompile(`application.name = "(.*)"`)
		regex_volume := regexp.MustCompile(`front-left: .+ (\d+)%`)

		sinkinput_num := si_num_r.FindString(first_line)
		app_name := regex_app_name.FindStringSubmatch(el)[1]
		vol_left := regex_volume.FindStringSubmatch(el)[1]

		si_info.Name = app_name
		si_info.Vol, _ = strconv.Atoi(vol_left)
		si_info.SI_number = sinkinput_num

		pulse_iface.sink_input_cache[sinkinput_num] = si_info
	}
}

func NewPulseCMDLineInteface() (pulse_iface *PulseCMDLineInterface) {
	pulse_iface = &PulseCMDLineInterface{}
	pulse_iface.build_sink_input_cache()

	return pulse_iface
}
