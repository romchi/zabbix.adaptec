package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type ldInfo struct {
	LdName              string `json:"logical device name"`
	BlockSize           string `json:"block size of member drives"`
	RaidLevel           string `json:"raid level"`
	UniqueIdentifier    string `json:"unique identifier"`
	StatusLD            string `json:"status of logical device"`
	Size                string `json:"size"`
	ParitySpace         string `json:"parity space"`
	StripeUnitSize      string `json:"stripe-unit size"`
	InterfaceType       string `json:"interface type"`
	DeviceType          string `json:"device type"`
	ReadCacheSettings   string `json:"read-cache setting"`
	ReadCacheStatus     string `json:"read-cache status"`
	WriteCacheSettings  string `json:"write-cache setting"`
	WriteCacheStatus    string `json:"write-cache status"`
	Partitioned         string `json:"partitioned"`
	ProtectedByHotSpare string `json:"protected by hot-spare"`
	Bootable            string `json:"bootable"`
	FailedStripes       string `json:"failed stripes"`
	PowerSettings       string `json:"power settings"`
}

func ldDiscovery() {
	deviceType := "LD"
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
			parts := strings.Split(string(out), "Logical Device number")
			for _, ldinfo := range parts {
				ld := discoveryDevice{}

				for _, ldstat := range strings.Split(ldinfo, "\n") {
					ld.ldDiscoveryParser(ldstat, controller)
				}
				if len(ld.Present) > 1 {
					disks = append(disks, ld)
				}
			}
		}
	}
	data := data{Data: disks}
	//r, _ := json.MarshalIndent(data, "", " ")
	r, _ := json.Marshal(data)
	fmt.Print(string(r))
}

func ldStats(ldName string) {
	deviceType := "LD"
	devices := map[string]ldInfo{}

	controllers, err := controllersCount()
	if err != nil {
		fmt.Print("Cannot check lspci adaptec controllers.")
		os.Exit(1)
	}
	bin, binErr := getBin("arcconf")
	if binErr != nil {
		fmt.Print("Arcconf not found.")
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
			ld := ldInfo{}

			for _, pdstat := range strings.Split(pdinfo, "\n") {
				ld.ldParserInfo(pdstat)
			}
			devices[ld.UniqueIdentifier] = ld
		}

		if _, ok := devices[ldName]; ok {
			//r, _ := json.MarshalIndent(devices[ldName], "", "  ")
			r, _ := json.Marshal(devices[ldName])
			fmt.Print(string(r))
		} else {
			fmt.Printf("LD not exist %v", ldName)
			os.Exit(1)
		}
	}
}

func (d *discoveryDevice) ldDiscoveryParser(line string, controller int) error {
	split := strings.Split(line, " : ")
	match := strings.ToLower(strings.TrimSpace(split[0]))
	if match == "logical device name" {
		d.DeviceAlias = strings.TrimSpace(split[1])
	}
	if match == "unique identifier" {
		d.DeviceID = strings.TrimSpace(split[1])
		d.Present = strings.TrimSpace(split[1])
	}
	d.DeviceType = "LD"
	return nil
}

func (ld *ldInfo) ldParserInfo(line string) error {
	split := strings.Split(line, " : ")
	match := strings.ToLower(strings.TrimSpace(split[0]))
	switch match {
	case "logical device name":
		ld.LdName = strings.TrimSpace(split[1])
	case "block size of member drives":
		ld.BlockSize = strings.TrimSpace(split[1])
	case "raid level":
		ld.RaidLevel = strings.TrimSpace(split[1])
	case "unique identifier":
		ld.UniqueIdentifier = strings.TrimSpace(split[1])
	case "status of logical device":
		ld.StatusLD = strings.TrimSpace(split[1])
	case "size":
		ld.Size = strings.TrimSpace(split[1])
	case "parity space":
		ld.ParitySpace = strings.TrimSpace(split[1])
	case "stripe-unit size":
		ld.StripeUnitSize = strings.TrimSpace(split[1])
	case "interface type":
		ld.InterfaceType = strings.TrimSpace(split[1])
	case "device type":
		ld.DeviceType = strings.TrimSpace(split[1])
	case "read-cache setting":
		ld.ReadCacheSettings = strings.TrimSpace(split[1])
	case "read-cache status":
		ld.ReadCacheStatus = strings.TrimSpace(split[1])
	case "write-cache setting":
		ld.WriteCacheSettings = strings.TrimSpace(split[1])
	case "write-cache status":
		ld.WriteCacheStatus = strings.TrimSpace(split[1])
	case "partitioned":
		ld.Partitioned = strings.TrimSpace(split[1])
	case "protected by hot-spare":
		ld.ProtectedByHotSpare = strings.TrimSpace(split[1])
	case "bootable":
		ld.Bootable = strings.TrimSpace(split[1])
	case "failed stripes":
		ld.FailedStripes = strings.TrimSpace(split[1])
	case "power settings":
		ld.PowerSettings = strings.TrimSpace(split[1])
	}
	return nil
}
