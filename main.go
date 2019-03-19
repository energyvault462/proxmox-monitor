package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var testMode bool

var upPercent, downPercent float64

func main() {

	var apcResult, serverIP, serverPort string
	var curVolts, curBattery float64
	var serverState, desiredServerState bool

	flag.BoolVar(&testMode, "t", false, "enable test mode")
	flag.Float64Var(&upPercent, "up", 50.0, "charge level of battery to reach before turning server on")
	flag.Float64Var(&downPercent, "down", 35.0, "charge level of battery to reach before shutdown of server")
	flag.StringVar(&serverIP, "ip", "192.168.1.2", "ip address of the server monoriting.")
	flag.StringVar(&serverPort, "port", "80", "port of the server monoriting.")

	flag.Parse()

	fmt.Printf("Testmode:     %v\nBattery up:   %v%%\nBattery down: %v%%\n", testMode, upPercent, downPercent)

	// start of loop
	apcResult = RunApcAccess()

	serverState = true // delete this when getServerState() is enabled.
	//serverState = getServerState

	curVolts, curBattery = getBatteryStats(apcResult)
	desiredServerState = GetDesiredPowerState(serverState, curBattery, curVolts)
	serverState = isServerRunning(serverIP, serverPort)
	if desiredServerState != serverState {
		// need to call code to shut down or turn on the server... unless this is in test mode
		fmt.Println("Server's power needs to be changed to", desiredServerState)
	}
}

// getBatteryStats parses a string in the APC "apcaccess" format
// and returns the line volt and time left of battery (in minutes) (float64)
func getBatteryStats(t string) (float64, float64) {
	volts := 0.0
	charge := 0.0

	re := regexp.MustCompile("LINEV\\s*.\\s*(\\d*.\\d*)")
	matches := re.FindStringSubmatch(t)
	if len(matches) > 1 {
		if s, err := strconv.ParseFloat(matches[1], 64); err == nil {
			volts = s
		} else {
			log.Println(err)
			os.Exit(1)
		}
	}

	re = regexp.MustCompile("BCHARGE\\s*.\\s*(\\d*.\\d*)")
	matches = re.FindStringSubmatch(t)
	if len(matches) > 1 {
		if s, err := strconv.ParseFloat(matches[1], 64); err == nil {
			charge = s
		} else {
			log.Println(err)
			os.Exit(1)
		}
	}
	return volts, charge
}

// RunApcAccess calls the apcaccess utility and returns its response.
func RunApcAccess() string {
	var cmd string
	if testMode == true {
		cmd = "../echoBattery/echoBattery.exe"
	} else {
		cmd = "apcaccess"
	}

	out, err := exec.Command(cmd).Output()
	if err != nil {
		log.Fatal(err)
	}

	s := string(out[:len(out)])
	return s
}

func GetDesiredPowerState(serverPower bool, curBattery, curVolt float64) bool {
	if serverPower == true && curVolt == 0 && curBattery <= downPercent {
		return false
	} else if serverPower == false && curVolt > 0 && curBattery >= upPercent {
		return true
	}

	return serverPower
}

func isServerRunning(site, port string) bool {
	timeout := time.Duration(1 * time.Second)
	_, err := net.DialTimeout("tcp", site+":"+port, timeout)
	if err != nil {
		fmt.Println("Site unreachable")
		return false
	}
	fmt.Println("site is working")
	return true

}
