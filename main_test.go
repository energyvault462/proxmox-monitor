package main

import "testing"

type ups struct {
	data        string
	parseAnswer []float64
}

var upsTests = []ups{
	ups{`CABLE    : USB Cable
	LINEV    : 119.0 Volts
	BCHARGE :  40.0 Percent
	`, []float64{119.0, 40.0}},

	ups{`CABLE    : USB Cable
	LINEV    : 119.0 Volts
	BCHARGE :  15.0 Percent
	`, []float64{119.0, 15.0}},

	ups{`CABLE    : USB Cable
	LINEV    : 119.0 Volts
	BCHARGE :  39.9 Percent
	`, []float64{119.0, 39.9}},

	ups{`CABLE    : USB Cable
	LINEV    : 0.0 Volts
	BCHARGE :  40.0 Percent
	`, []float64{0.0, 40.0}},

	ups{`CABLE    : USB Cable
	LINEV    : 0.0 Volts
	BCHARGE :  15.0 Percent
	`, []float64{0.0, 15.0}},

	ups{`CABLE    : USB Cable
	LINEV    : 0.0 Volts
	BCHARGE :  39.9 Percent
	`, []float64{0.0, 39.9}},
}

type serverState struct {
	batteryUp     float64
	batteryDown   float64
	serverOn      bool
	batteryLevel  float64
	volt          float64
	desiredServer bool // should the server be on or not
}

var serverStateTests = []serverState{
	serverState{75.0, 40.0, true, 40.1, 0, true},
	serverState{75.0, 40.0, true, 40.1, 119.0, true},
	serverState{75.0, 40.0, true, 40.0, 0, false},
	serverState{75.0, 40.0, true, 40.0, 119.0, true},
	serverState{75.0, 40.0, false, 75.0, 0, false},
	serverState{75.0, 40.0, false, 75.0, 119.0, true},
	serverState{75.0, 40.0, false, 74.9, 0, false},
	serverState{75.0, 40.0, false, 74.9, 119.0, false},
}

func TestGetBatteryStats(t *testing.T) {

	for _, v := range upsTests {
		volt, p := getBatteryStats(v.data)
		if volt != v.parseAnswer[0] || p != v.parseAnswer[1] {
			t.Errorf("getBatteryStats failed: expected %v volts and %v, got %v and %v\n", v.parseAnswer[0], v.parseAnswer[1], volt, p)
		}
	}
}

func TestGetDesiredPowerState(t *testing.T) {

	for _, v := range serverStateTests {
		upPercent = v.batteryUp
		downPercent = v.batteryDown
		r := GetDesiredPowerState(v.serverOn, v.batteryLevel, v.volt)
		if r != v.desiredServer {
			t.Errorf("GetDesiredPowerState failed: Expected %v, got %v\nServer State: %v\n\tBattery Up Set Level: %v%%\n\tBattery Down Set Level: %v%%\n\tBattery Level: %v%%\n\tVolts: %v\n", v.desiredServer, r, v.serverOn, v.batteryUp, v.batteryDown, v.batteryLevel, v.volt)
		}
	}
}
