package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type adInfo struct {
	ControllerStatus           string `json:"controller status"`
	ChannelDescription         string `json:"channel description"`
	ControllerModel            string `json:"controller model"`
	ControllerSerialNumber     string `json:"controller serial number"`
	ControllerWorldWideName    string `json:"controller world wide name"`
	ControllerAlarm            string `json:"controller alarm"`
	Temperature                string `json:"temperature"`
	InstalledMemory            string `json:"installed memory"`
	GlobalTaskPriority         string `json:"global task priority"`
	PerformanceMode            string `json:"performance mode"`
	StayawakePeriod            string `json:"stayawake period"`
	DefunctDiskDriveCount      int    `json:"defunct disk drive count"`
	LogicalDevicesFailed       int    `json:"logical devices failed"`
	LogicalDevicesTotal        int    `json:"logical devices total"`
	LogicalDevicesDegraded     int    `json:"logical devices degraded"`
	NCQStatus                  string `json:"ncq status"`
	Copyback                   string `json:"copyback"`
	AutomaticFailover          string `json:"automatic failover"`
	BackgroundConsistencyCheck string `json:"background consistency check"`
	BIOS                       string `json:"bios"`
	Firmware                   string `json:"firmware"`
	Driver                     string `json:"driver"`
	Status                     string `json:"status"`
	BatteryPresent             string `json:"battery present"`
}

func adDiscovery() {
	deviceType := "AD"
	adapters := []discoveryDevice{}

	controllers, err := controllersCount()
	if err != nil {
		fmt.Printf("Cannot check lspci adaptec controllers\n - %v", err)
		os.Exit(1)
	}

	if controllers > 0 {
		bin, binErr := getBin("arcconf")
		if binErr != nil {
			fmt.Printf("arcconf - not found in system PATH\n - %v", err)
			os.Exit(0)
		}

		for controller := 1; controller <= controllers; controller++ {
			args := []string{"getconfig", strconv.Itoa(controller), deviceType, "nologs"}
			cmd := exec.Command(bin, args...)
			out, err := cmd.Output()
			if err != nil {
				fmt.Printf("Error %v", err)
			}

			ad := discoveryDevice{}
			for _, adstat := range strings.Split(string(out), "\n") {
				ad.adDiscoveryParser(adstat, controller)
			}
			adapters = append(adapters, ad)
		}
	}
	data := data{Data: adapters}

	//r, _ := json.MarshalIndent(data, "", "  ")
	r, _ := json.Marshal(data)
	fmt.Print(string(r))
}

func adStats(adController string) {
	deviceType := "AD"

	controllers, cErr := controllersCount()
	if cErr != nil {
		fmt.Printf("Cannot check lspci adaptec controllers: %v", cErr)
		os.Exit(1)
	}

	bin, binErr := getBin("arcconf")
	if binErr != nil {
		fmt.Println("Arcconf not found:", binErr)
		os.Exit(0)
	}

	ads := map[string]adInfo{}

	for controller := 1; controller <= controllers; controller++ {
		args := []string{"getconfig", strconv.Itoa(controller), deviceType, "nologs"}
		cmd := exec.Command(bin, args...)
		out, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error %v", err)
		}

		ad := adInfo{Status: "NotPresent"}

		for _, adstat := range strings.Split(string(out), "\n") {
			ad.adParserInfo(adstat)
		}

		ads[strconv.Itoa(controller)] = ad
	}

	if _, ok := ads[adController]; ok {
		//r, _ := json.MarshalIndent(devices[ldName], "", "  ")
		r, _ := json.Marshal(ads[adController])
		fmt.Print(string(r))
	} else {
		fmt.Printf("AD not exist %v", adController)
		os.Exit(1)
	}
}

func (d *discoveryDevice) adDiscoveryParser(line string, controller int) error {
	split := strings.Split(line, " : ")
	match := strings.ToLower(strings.TrimSpace(split[0]))
	if match == "controller model" {
		model := strings.TrimSpace(split[1])
		d.DeviceAlias = model
	}
	d.DeviceID = strconv.Itoa(controller)
	d.Present = "Present"
	d.DeviceType = "AD"
	return nil
}

func (ad *adInfo) adParserInfo(line string) error {
	split := strings.Split(line, " : ")
	match := strings.ToLower(strings.TrimSpace(split[0]))
	switch match {
	case "controller status":
		ad.ControllerStatus = strings.TrimSpace(split[1])
	case "channel description":
		ad.ChannelDescription = strings.TrimSpace(split[1])
	case "controller model":
		ad.ControllerModel = strings.TrimSpace(split[1])
	case "controller serial number":
		ad.ControllerSerialNumber = strings.TrimSpace(split[1])
	case "controller world wide name":
		ad.ControllerWorldWideName = strings.TrimSpace(split[1])
	case "controller alarm":
		ad.ControllerAlarm = strings.TrimSpace(split[1])
	case "temperature":
		ad.Temperature = strings.TrimSpace(split[1])
	case "installed memory":
		ad.InstalledMemory = strings.TrimSpace(split[1])
	case "global task priority":
		ad.GlobalTaskPriority = strings.TrimSpace(split[1])
	case "performance mode":
		ad.PerformanceMode = strings.TrimSpace(split[1])
	case "stayawake period":
		ad.StayawakePeriod = strings.TrimSpace(split[1])
	case "defunct disk drive count":
		ad.DefunctDiskDriveCount = 0
	case "logical devices/failed/degraded":
		device := strings.Split(split[1], "/")
		var r []int
		for _, i := range device {
			d, err := strconv.Atoi(i)
			if err != nil {
				return err
			}
			r = append(r, d)
		}
		ad.LogicalDevicesTotal = r[0]
		ad.LogicalDevicesFailed = r[1]
		ad.LogicalDevicesDegraded = r[2]
	case "ncq status":
		ad.NCQStatus = strings.TrimSpace(split[1])
	case "copyback":
		ad.Copyback = strings.TrimSpace(split[1])
	case "automatic failover":
		ad.AutomaticFailover = strings.TrimSpace(split[1])
	case "background consistency check":
		ad.BackgroundConsistencyCheck = strings.TrimSpace(split[1])
	case "bios":
		ad.BIOS = strings.TrimSpace(split[1])
	case "firmware":
		ad.Firmware = strings.TrimSpace(split[1])
	case "driver":
		ad.Driver = strings.TrimSpace(split[1])
	case "status":
		ad.Status = strings.TrimSpace(split[1])
	case "controller zmm information":
		ad.BatteryPresent = "True"
	case "battery":
		ad.BatteryPresent = "True"
	}
	return nil
}
