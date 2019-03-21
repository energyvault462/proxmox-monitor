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

	"github.com/linde12/gowol"
)

var testMode bool

var upPercent, downPercent float64

func main() {

	var apcResult, serverIP, serverPort, macAddress, userName string
	var curVolts, curBattery float64
	var serverState, desiredServerState bool

	flag.BoolVar(&testMode, "t", false, "enable test mode: -t")
	flag.Float64Var(&upPercent, "up", 50.0, "charge level of battery to reach before turning server on: -up=\"50.0\"")
	flag.Float64Var(&downPercent, "down", 35.0, "charge level of battery to reach before shutdown of server: -down=\"35.0\"")
	flag.StringVar(&serverIP, "ip", "192.168.1.2", "ip address of the server monoriting: -ip=\"192.168.1.2\"")
	flag.StringVar(&serverPort, "port", "80", "port of the server monoriting: -port=\"80\"")
	flag.StringVar(&userName, "user", "root", "user to ssh into the server for shutdown: -user=\"root\"")
	flag.StringVar(&macAddress, "mac", "AA:AA:AA:AA:AA:AA", "MAC address of the server to send a WOL magic packet:  -mac=\"AA:AA:AA:AA:AA:AA\"")

	flag.Parse()

 	file, err := os.OpenFile("/logs/pmm.log",  os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	log.SetOutput(file)
	log.Printf("Starting proxmox-monitor\n\tServerIP:     %v\n\tTestmode:     %v\n\tBattery up:   %v%%\n\tBattery down: %v%%\n", serverIP, testMode, upPercent, downPercent)

	// start of loop
	for true {
		apcResult = RunApcAccess()

		serverState = true // delete this when getServerState() is enabled.
		//serverState = getServerState

		curVolts, curBattery = getBatteryStats(apcResult)
		serverState = isServerRunning(serverIP, serverPort)
		log.Printf(GetStatusStr(curVolts, curBattery, serverState))
		desiredServerState = GetDesiredPowerState(serverState, curBattery, curVolts)
		if desiredServerState != serverState {
			// need to call code to shut down or turn on the server... unless this is in test mode
			log.Println("Server's power needs to be changed to", desiredServerState)
			if desiredServerState == true {
				// send magic packet to turn it on
				log.Println("Power is on, battery above accpetable limit. Sending WOL to server...")
				if packet, err := gowol.NewMagicPacket(macAddress); err == nil {
					packet.Send("255.255.255.255")          // send to broadcast
					packet.SendPort("255.255.255.255", "7") // specify receiving port
				}
				time.Sleep(5 * time.Minute)
			} else {
				// ssh to server and send WOL
				// TODO: switch this to an api call to the server.
				log.Println("Power is off, battery dropped below accpetable limit. Sending shutdown to server...")
				ShutdownViaSSH(userName, serverIP)
				time.Sleep(5 * time.Minute)
			}
		} else {
			time.Sleep(1 * time.Minute)
		}

	}

	log.Println("Finishing proxmox-monitor")
}

// LogAndNotify sends messages to logs as well as pushover.
func LogAndNotify(file *os.File) {

}

// GetStatusStr returns a string of the various statuses. Can be used for logging and debuggingf
func GetStatusStr(curVolts, curBattery float64, serverOn bool) string {
	s := fmt.Sprintf("Server on: %v | Current Volts: %3.1f | Current Battery: %3.1f | Shutdown at: %3.1f | Startup at: %3.1f", serverOn, curVolts, curBattery, downPercent, upPercent)
	return s
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

// ShutdownViaSSH uses ssh name@ip to issue a shutdown command.
func ShutdownViaSSH(name, ip string) {
	server := fmt.Sprintf("%v@%v", name, ip)
	log.Printf("   Running: ssh %v sudo /sbin/shutdown\n", server)
	out, err := exec.Command("ssh", server, "sudo /sbin/shutdown").Output()
	if err != nil {
		log.Printf("   error in shutdown.\n   return: %v\n   error: %v\n", out, err)
	}
}

// RunApcAccess calls the apcaccess utility and returns its response.
func RunApcAccess() string {
	var cmd string
	if testMode == true {
		cmd = "../../../echoBattery/echoBattery.exe"
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
		//log.Println("Site unreachable")
		return false
	}
	//log.Println("site is working")
	return true

}
