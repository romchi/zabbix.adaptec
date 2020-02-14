package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type pdInfo struct {
	DeviceID             string `json:"device id"`
	State                string `json:"state"`
	BlockSize            string `json:"block size"`
	Supported            string `json:"supported"`
	TransferSpeed        string `json:"transfer speed"`
	Vendor               string `json:"vendor"`
	Model                string `json:"model"`
	Firmware             string `json:"firmware"`
	SerialNumber         string `json:"serial number"`
	ReservedSize         string `json:"reserved size"`
	UsedSize             string `json:"used size"`
	UnusedSize           string `json:"unused size"`
	TotalSize            string `json:"total size"`
	WriteCache           string `json:"write cache"`
	FRU                  string `json:"fru"`
	Smart                string `json:"s.m.a.r.t."`
	SmartWarnings        int    `json:"s.m.a.r.t. warnings"`
	PowerState           string `json:"power state"`
	SupportedPowerStates string `json:"supported power state"`
	SSD                  string `json:"ssd"`
	NCQ                  string `json:"ncq"`
}

func pdDiscovery() {
	deviceType := "PD"
	disks := []discoveryDevice{}

	controllers, err := controllersCount()
	if err != nil {
		fmt.Printf("Cannot check lspci adaptec controllers\n - %v", err)
		os.Exit(1)
	}

	if controllers > 0 {
		bin, binErr := getBin("arcconf")
		if binErr != nil {
			fmt.Printf("arcconf - not found in system PATH\n - %v", binErr)
			os.Exit(0)
		}

		for controller := 1; controller <= controllers; controller++ {
			args := []string{"getconfig", strconv.Itoa(controller), deviceType, "nologs"}
			cmd := exec.Command(bin, args...)
			out, err := cmd.Output()
			if err != nil {
				fmt.Printf("Error %v", err)
			}
			parts := strings.Split(string(out), "Device #")
			for _, pdinfo := range parts {
				pd := discoveryDevice{}

				for _, pdstat := range strings.Split(pdinfo, "\n") {
					pd.pdDiscoveryParser(pdstat, controller)
				}
				if len(pd.Present) > 1 {
					disks = append(disks, pd)
				}
			}
		}
	}
	data := data{Data: disks}

	//r, _ := json.MarshalIndent(data, "", " ")
	r, _ := json.Marshal(data)
	fmt.Print(string(r))
}

func pdStats(pdName string) {
	deviceType := "PD"

	controllers, err := controllersCount()
	if err != nil {
		fmt.Printf("Cannot check lspci adaptec controllers: %v", err)
		os.Exit(1)
	}

	bin, binErr := getBin("arcconf")
	if binErr != nil {
		fmt.Println("Arcconf not found:", err)
		os.Exit(0)
	}

	for controller := 1; controller <= controllers; controller++ {
		args := []string{"getconfig", strconv.Itoa(controller), deviceType, "nologs"}
		cmd := exec.Command(bin, args...)
		out, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error %v", err)
		}

		parts := strings.Split(string(out), "Device #")
		disk := map[string]pdInfo{}

		for _, pdinfo := range parts {
			pd := pdInfo{}

			for _, pdstat := range strings.Split(pdinfo, "\n") {
				pd.pdParserInfo(pdstat, controller)
			}
			if len(pd.State) > 1 {
				disk[pd.DeviceID] = pd
			}
		}

		if _, ok := disk[pdName]; ok {
			//r, _ := json.MarshalIndent(disk[pdName], "", " ")
			r, _ := json.Marshal(disk[pdName])
			fmt.Print(string(r))
		} else {
			fmt.Printf("PD not exist %v", pdName)
			os.Exit(1)
		}
	}
}

func (d *discoveryDevice) pdDiscoveryParser(line string, controller int) error {
	split := strings.Split(line, " : ")
	match := strings.ToLower(strings.TrimSpace(split[0]))
	if match == "reported location" {
		text := "Controller " + strconv.Itoa(controller) + ", " + strings.TrimSpace(split[1])
		d.DeviceID = text
	}
	if match == "state" {
		state := strings.TrimSpace(split[1])
		d.Present = state
	}
	d.DeviceType = "PD"
	return nil
}

func (pd *pdInfo) pdParserInfo(line string, controller int) error {
	split := strings.Split(line, " : ")
	match := strings.ToLower(strings.TrimSpace(split[0]))
	switch match {
	case "reported location":
		text := "Controller " + strconv.Itoa(controller) + ", " + strings.TrimSpace(split[1])
		pd.DeviceID = text
	case "state":
		pd.State = strings.TrimSpace(split[1])
	case "block size":
		pd.BlockSize = strings.TrimSpace(split[1])
	case "supported":
		pd.Supported = strings.TrimSpace(split[1])
	case "transfer speed":
		pd.TransferSpeed = strings.TrimSpace(split[1])
	case "vendor":
		pd.Vendor = strings.TrimSpace(split[1])
	case "model":
		pd.Model = strings.TrimSpace(split[1])
	case "firmware":
		pd.Firmware = strings.TrimSpace(split[1])
	case "serial number":
		pd.SerialNumber = strings.TrimSpace(split[1])
	case "reserved size":
		pd.ReservedSize = strings.TrimSpace(split[1])
	case "used size":
		pd.UsedSize = strings.TrimSpace(split[1])
	case "unused size":
		pd.UnusedSize = strings.TrimSpace(split[1])
	case "total size":
		pd.TotalSize = strings.TrimSpace(split[1])
	case "write cache":
		pd.WriteCache = strings.TrimSpace(split[1])
	case "fru":
		pd.FRU = strings.TrimSpace(split[1])
	case "s.m.a.r.t.":
		pd.Smart = strings.TrimSpace(split[1])
	case "s.m.a.r.t. warnings":
		res, err := strconv.Atoi(strings.TrimSpace(split[1]))
		if err != nil {
			return err
		}
		pd.SmartWarnings = res
	case "power state":
		pd.PowerState = strings.TrimSpace(split[1])
	case "supported power state":
		pd.SupportedPowerStates = strings.TrimSpace(split[1])
	case "ssd":
		pd.SSD = strings.TrimSpace(split[1])
	case "ncq status":
		pd.NCQ = strings.TrimSpace(split[1])
	}
	return nil
}
