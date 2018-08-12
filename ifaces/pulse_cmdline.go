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
	"io/ioutil"
	"encoding/base64"
)

type sinkInputInfo struct {
	Name      string
	Vol       int
	SI_number string
	Icon	  string
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
			vol_regex_res := regex_volume.FindStringSubmatch(el)
			vol_left := "0"
			if len(vol_regex_res) >= 2 {
				vol_left = regex_volume.FindStringSubmatch(el)[1]
			}
			vol, err := strconv.Atoi(vol_left)
			return vol, err
		}
	}

	return -1, errors.New("Cannot get volume")
}

func si_details_to_update(si_info sinkInputInfo) (upd VoleurUpdateType) {
	upd.Name = si_info.Name
	upd.Type = AddOrUpdate
	upd.Val = si_info.Vol
	if upd.AuxData == nil {
		upd.AuxData = make(map[string]string)
	}
	upd.UID = si_info.SI_number
	upd.AuxData["icon"] = si_info.Icon

	return upd
}

func (pulse_iface *PulseCMDLineInterface) get_updated_sinkinput_details(str string) (VoleurUpdateType, error) {
	r := regexp.MustCompile(`[\d]+`)

	sinkinput_num := r.FindString(str)
	cached_details, ok := pulse_iface.sink_input_cache[sinkinput_num]

	if !ok {
		return VoleurUpdateType{}, errors.New("Failed to get sinkinput details")
	}

	out := si_details_to_update(cached_details)
	out.AuxData["isSinkVol"] = "0"

	vol, err := pactl_get_sinkinput_volume(sinkinput_num)
	out.Val = vol

	if err != nil {
		return VoleurUpdateType{}, err
	} else {
		return out, nil
	}
}

func (pulse_iface *PulseCMDLineInterface) parse_event(str string) (VoleurUpdateType, error) {
	var out VoleurUpdateType
	var err error

	// TODO handle sink volume
	if strings.Contains(str, "'change' on sink-input") {
		out, err = pulse_iface.get_updated_sinkinput_details(str)
		if err != nil {
			return VoleurUpdateType{}, err
		}
		out.Type = AddOrUpdate

		// Update sink_input cache
		bk, ok := pulse_iface.sink_input_cache[out.UID]
		if !ok {
			return VoleurUpdateType{}, err
		}
		bk.Vol = out.Val
		pulse_iface.sink_input_cache[out.UID] = bk

	} else if strings.Contains(str, "'new' on sink-input") {
		fmt.Println("new si")
		// TODO optimise?
		pulse_iface.build_sink_input_cache()

		r := regexp.MustCompile(`[\d]+`)
		sinkinput_num := r.FindString(str)
		cached_details, ok := pulse_iface.sink_input_cache[sinkinput_num]

		if !ok {
			return VoleurUpdateType{}, errors.New("Failed to get sinkinput details")
		}

		out := si_details_to_update(cached_details)
		out.Type = AddOrUpdate

		return out, nil

	} else if strings.Contains(str, "'remove' on sink-input") {
		// TODO optimise?
		fmt.Println("si removed")
		r := regexp.MustCompile(`[\d]+`)
		sinkinput_num := r.FindString(str)
		cached_details, ok := pulse_iface.sink_input_cache[sinkinput_num]

		if !ok {
			return VoleurUpdateType{}, errors.New("Failed to get sinkinput details")
		}

		out := si_details_to_update(cached_details)
		out.Type = Remove

		pulse_iface.build_sink_input_cache()

		return out, nil
	} else {
		return VoleurUpdateType{}, errors.New("Not the update you're looking for")
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
		vol_to_set := (PA_VOLUME_MAX * update.Val) / 100
		fmt.Println("Will apply ", update)
		si_number := update.UID

		if _, ok := pulse_iface.sink_input_cache[si_number]; !ok {
			fmt.Println("Invalid SI number")
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

var include_icon = true

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

		app_name := ""
		vol_left := "0"

		app_name_res := regex_app_name.FindStringSubmatch(el)
		if len(app_name_res) < 2 {
			regex_media_name := regexp.MustCompile(`media.name = "(.*)"`)
			media_name_res := regex_media_name.FindStringSubmatch(el)
			if len(media_name_res) >= 2 {
				app_name = media_name_res[1]
			}
		} else {
			app_name = app_name_res[1]
		}

		vol_left_res := regex_volume.FindStringSubmatch(el)
		if len(vol_left_res) >= 2 {
			vol_left = regex_volume.FindStringSubmatch(el)[1]
		}

		if include_icon {
			//		application.icon_name = "spotify-client"
			regex_icon_name := regexp.MustCompile(`application.icon_name = "(.*)"`)
	
			icon_name_res := regex_icon_name.FindStringSubmatch(el)
			if len(icon_name_res) >= 2 {
				icon_name := icon_name_res[1]
//				fmt.Println("Icon name " + icon_name)
				icon_path := get_icon_path(icon_name)
				fmt.Println("Icon path " + icon_path)
				si_info.Icon = get_base64_file(icon_path)
			}
		}
		
		si_info.Name = app_name
		si_info.Vol, _ = strconv.Atoi(vol_left)
		si_info.SI_number = sinkinput_num

		pulse_iface.sink_input_cache[sinkinput_num] = si_info
	}
}

func get_base64_file(path string) (enc string) {
	buff, err := ioutil.ReadFile(strings.TrimSpace(path))
	if err != nil {
		fmt.Println("err:", err)
		return ""
	}

	return base64.StdEncoding.EncodeToString(buff)
}

func get_icon_path(name string) (path string) {
	cmd_out, err := exec.Command("./src/voleur/aux/get_icon_path_gtk.py", name).Output()
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
	
	return string(cmd_out)
}

func (pulse_iface *PulseCMDLineInterface) GetAll() (ar_json [][]byte) {
	for _, sinkinfo := range pulse_iface.sink_input_cache {
		update := si_details_to_update(sinkinfo)
		//		update.Name = "this is a super long name to test auto sizing"
		b, err := json.Marshal(update)
		if err != nil {
			continue
		}
		ar_json = append(ar_json, b)
	}

	return
}

func NewPulseCMDLineInteface() (pulse_iface *PulseCMDLineInterface) {
	pulse_iface = &PulseCMDLineInterface{}
	pulse_iface.build_sink_input_cache()

	return pulse_iface
}
