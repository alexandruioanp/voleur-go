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

type PulseCMDLineInterface struct {
}

const PA_VOLUME_MAX = 65536

func (pulse_iface *PulseCMDLineInterface) Listen(json_out chan []byte) {
	change_chan := make(chan string)
	go pulse_iface.listen_pulse(change_chan)
	go decode(change_chan, json_out)
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

func pactl_get_sinkinput_details(sinkinput_num string) (VoleurUpdateType, error) {
	// TODO cache these?
	cmd_out, err := exec.Command("pactl", "list", "sink-inputs").Output()
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}

	out := VoleurUpdateType{Name: "", Vol: 0, IsSinkVol: false}

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
			out.Name = app_name
			out.Vol, _ = strconv.Atoi(vol_left)
			out.Type = Update
			out.si_number = "0"
			// found sinkinput #sinkinput_num
		}

	}

	return out, nil
}

func parse_pactl_update_msg(str string) (VoleurUpdateType, error) {
	var out VoleurUpdateType

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
		str := <-change_in
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

func (pulse_iface *PulseCMDLineInterface) ApplyChanges(in_chan chan VoleurUpdateType) {
	for {
		update := <-in_chan
		vol_to_set := (PA_VOLUME_MAX * update.Vol) / 100
		update.si_number = "0"
		fmt.Printf("Setting volume %d: %s to %s\n", vol_to_set, strconv.Itoa(vol_to_set), update.si_number)
		cmd := exec.Command("pactl", "set-sink-input-volume", update.si_number, strconv.Itoa(vol_to_set))
		//	pactl set-sink-input-volume 0 20000

		var stderr bytes.Buffer
		//	cmd.Stdout = &out
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error running Cmd", stderr.String())
			//		os.Exit(1)
		}
	}
}