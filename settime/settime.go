package settime

import (
	"github.com/pavelanni/tinygo-drivers/rotaryencoder"
	"github.com/pavelanni/tinygo-drivers/tm1637"
)

type TimeSetState struct {
	hourIsSet   bool
	minuteIsSet bool
}

// timeSetStates is an array of states which has three elements
// - both hour and minute are set
// - hour is not set, minute is set
// - hour is set, minute is not set
// Each click of the switch changes from one state to the next
var TimeSetStates = []TimeSetState{
	TimeSetState{
		hourIsSet:   true,
		minuteIsSet: true,
	},
	TimeSetState{
		hourIsSet:   false,
		minuteIsSet: true,
	},
	TimeSetState{
		hourIsSet:   true,
		minuteIsSet: false,
	},
}

func SetTime(enc *rotaryencoder.Device,
	display *tm1637.Device,
	hour *uint8,
	minute *uint8,
	tss *TimeSetState) {
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

func SetTimeState(enc *rotaryencoder.Device,
	tss *TimeSetState) {
	var curStateIndex int
	for {
		if <-enc.Switch {
			curStateIndex++
			i := curStateIndex % len(TimeSetStates) // cicrular around the array
			*tss = TimeSetStates[i]
		}
	}
}
