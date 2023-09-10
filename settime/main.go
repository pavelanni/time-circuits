package main

import (
	"machine"

	"github.com/pavelanni/tinygo-drivers/rotaryencoder"
	"github.com/pavelanni/tinygo-drivers/tm1637"
)

type timeSetState struct {
	hourIsSet   bool
	minuteIsSet bool
}

// timeSetStates is an array of states which has three elements
// - both hour and minute are set
// - hour is not set, minute is set
// - hour is set, minute is not set
// Each click of the switch changes from one state to the next
var timeSetStates = []timeSetState{
	timeSetState{
		hourIsSet:   true,
		minuteIsSet: true,
	},
	timeSetState{
		hourIsSet:   false,
		minuteIsSet: true,
	},
	timeSetState{
		hourIsSet:   true,
		minuteIsSet: false,
	},
}

func setTime(enc *rotaryencoder.Device,
	display *tm1637.Device,
	hour *uint8,
	minute *uint8,
	tss *timeSetState) {
	display.DisplayClock(*hour, *minute, true)
	for {
		delta := <-enc.Dir
		if !tss.hourIsSet {
			*hour = (*hour + uint8(delta) + 24) % 24
		} else if !tss.minuteIsSet {
			*minute = (*minute + uint8(delta) + 60) % 60
		}
		display.DisplayClock(*hour, *minute, true)
	}
}

func setTimeState(enc *rotaryencoder.Device,
	tss *timeSetState) {
	var curStateIndex int
	for {
		if <-enc.Switch {
			curStateIndex++
			i := curStateIndex % len(timeSetStates) // cicrular around the array
			*tss = timeSetStates[i]
		}
	}
}

func main() {
	emptyChan := make(chan bool)

	hour := uint8(0)
	minute := uint8(0)
	tss := timeSetStates[0]
	timeEnc := rotaryencoder.New(machine.GP7, machine.GP6, machine.GP8)
	timeEnc.Configure()

	timeDisplay := tm1637.New(machine.GP10, machine.GP11, 7) // clk, dio, brightness
	timeDisplay.Configure()
	timeDisplay.ClearDisplay()

	go setTimeState(&timeEnc, &tss)
	go setTime(&timeEnc, &timeDisplay, &hour, &minute, &tss)

	for {
		<-emptyChan
	}
}
