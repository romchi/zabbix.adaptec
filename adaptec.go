package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type data struct {
	Data []discoveryDevice `json:"data"`
}

type discoveryDevice struct {
	DeviceID    string `json:"{#DEVICE_ID}"`
	DeviceType  string `json:"{#DEVICE_TYPE}"`
	DeviceAlias string `json:"{#DEVICE_ALIAS}"`
	Present     string `json:"{#PRESENT}"`
}

func main() {
	discoveryCommand := flag.NewFlagSet("discover", flag.ExitOnError)
	statsCommand := flag.NewFlagSet("stats", flag.ExitOnError)

	discoveryDeviceType := discoveryCommand.String("type", "", "device type {ad, ld, pd} (Required)")

	statsDeviceType := statsCommand.String("type", "", "device type {ad, ld, pd} (Required)")
	statsDeviceName := statsCommand.String("name", "", `Device "name" to get stats (Required)`)

	if len(os.Args) < 2 {
		fmt.Println("[discovery, stats, check] - required one command")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "discovery":
		discoveryCommand.Parse(os.Args[2:])
	case "stats":
		statsCommand.Parse(os.Args[2:])
	case "check":
		checkArcconf()
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if discoveryCommand.Parsed() {
		metricChoices := map[string]bool{"ad": true, "ld": true, "pd": true}
		if _, validChoice := metricChoices[*discoveryDeviceType]; !validChoice {
			discoveryCommand.PrintDefaults()
			os.Exit(1)
		}
		switch *discoveryDeviceType {
		case "ad":
			adDiscovery()
		case "ld":
			ldDiscovery()
		case "pd":
			pdDiscovery()
		default:
			discoveryCommand.PrintDefaults()
			os.Exit(1)
		}
	}

	if statsCommand.Parsed() {
		metricChoices := map[string]bool{"ad": true, "ld": true, "pd": true}
		if _, validChoice := metricChoices[*statsDeviceType]; !validChoice {
			statsCommand.PrintDefaults()
			os.Exit(0)
		}
		if len(*statsDeviceName) < 1 {
			statsCommand.PrintDefaults()
			os.Exit(0)
		}
		switch *statsDeviceType {
		case "ad":
			adStats(*statsDeviceName)
		case "ld":
			ldStats(*statsDeviceName)
		case "pd":
			pdStats(*statsDeviceName)
		default:
			statsCommand.PrintDefaults()
			os.Exit(0)
		}
	}
}

func noDevice() {
	null := []discoveryDevice{}
	result := data{Data: null}
	r, _ := json.Marshal(result)
	fmt.Print(string(r))
}

func checkArcconf() {
	controllers, err := controllersCount()
	if err != nil {
		fmt.Printf("Cannot check lspci adaptec controllers\n - %v", err)
		os.Exit(1)
	}

	if controllers > 0 {
		_, binErr := getBin("arcconf")
		if binErr != nil {
			fmt.Print(0)
		}
		fmt.Print(1)
	} else {
		fmt.Print(1)
	}
}

func controllersCount() (int, error) {
	counter := 0
	bin, getErr := getBin("lspci")
	if getErr != nil {
		return counter, getErr
	}
	cmd := exec.Command(bin)
	out := &bytes.Buffer{}
	cmd.Stdout = out
	err := cmd.Run()
	if err != nil {
		return counter, err
	}
	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		//if strings.Contains(scanner.Text(), "Adaptec") {
		if strings.Contains(scanner.Text(), "Adaptec") && strings.Contains(scanner.Text(), "RAID bus controller") {
			counter++
		}
	}
	return counter, nil
}

func getBin(binFile string) (string, error) {
	location := []string{"/bin", "/sbin", "/usr/bin", "/usr/sbin", "/usr/local/bin", "/usr/local/sbin"}

	for _, path := range location {
		lookup := path + "/" + binFile
		fileInfo, err := os.Stat(path + "/" + binFile)
		if err != nil {
			continue
		}
		if !fileInfo.IsDir() {
			return lookup, nil
		}
	}
	return "", fmt.Errorf("Not found: '%v'", binFile)
}
