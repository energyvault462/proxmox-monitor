# proxmox-monitor
GoLang application to monitor battery level and shutdown/start proxmox server using a Raspberry Pi

This is a personal project with a side benefit to get more practice with GoLang. Instead of having the battery backup connected directly to the Proxmox server, it's connected to a Raspberry Pi.  This application will monitor the UPS and when power is out and battery drops below a defined level, it will issue a shutdown command to the server. When power is restored and batter level is high enough, it will send a WOL magic packet to turn on the Proxmox server. Goal is to run this in a Docker container on the Pi.

To run:
  go build
  ./proxmox-monitor -ip="192.168.1.2" -down=35.0 -up=20.0


./proxmox-monitor -h    will show information on the available flags.

TODO:
  - issue shutdown command to server via ssh
    - Later goal is to make a REST service that runs on the Proxmox server to receive the shutdown command and do so.
  - send WOL command
  - put in a loop so it continues to run
  - logging
  - add Pushover support
  - Create Docker container for deployment
