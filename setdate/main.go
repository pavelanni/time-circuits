package main

import (
	"machine"

	"github.com/pavelanni/tinygo-drivers/rotaryencoder"
	"github.com/pavelanni/tinygo-drivers/tm1637"
)

type dateSetState struct {
	monthIsSet bool
	dayIsSet   bool
}

// Yes, it's silly to list sequential numbers, but in general these slices can include month names, weekday names, etc.
// As we circle around the slice indices, we can use them with any slices.
var months = []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
var days = []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
var daysInMonth = []int{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

// dateSetStates is a slice of setting states which has three elements
// - both month and day are set
// - month is not set, day is set
// - month is set, day is not set
// Each click of the switch changes from one state to the next
var dateSetStates = []dateSetState{
	dateSetState{
		monthIsSet: true,
		dayIsSet:   true,
	},
	dateSetState{
		monthIsSet: false,
		dayIsSet:   true,
	},
	dateSetState{
		monthIsSet: true,
		dayIsSet:   false,
	},
}

// capIdx calculates the new index based on the given index, the limit (cap), and overflow flag.
// If over is true and if the index is greater than the cap, the new index is calculated as the
// remainder of the index divided by the cap. I.e. if the index is 6 and cap is 5, the new index is 1.
// The index circles around staying with the range.
// If over is false then the index is clamped to the range.
// I.e. if the index is 6 and cap is 5, the new index is 5.
// If the index is less than 0, the new index is 0.
//
// Parameters:
// - idx: the original index.
// - cap: the index limit.
// - over: the overflow flag.
//
// Returns:
// - newidx: the new index.
func capIdx(idx int, cap int, over bool) int {
	var newidx int
	if over {
		newidx = (idx%cap + cap) % cap
	} else {
		if idx > cap {
			newidx = cap
		} else if idx < 0 {
			newidx = 0
		} else {
			newidx = idx
		}
	}
	return newidx
}

// setDate updates the display with the current date based on user input from the rotary encoder.
//
// Parameters:
// - enc: A pointer to the rotary encoder device.
// - display: A pointer to the tm1637 device used for displaying the date.
// - monthIdx: A pointer to the index of the current month.
// - dayIdx: A pointer to the index of the current day.
// - dss: A pointer to the dateSetState object that keeps track of which part of the date is being set.
func setDate(enc *rotaryencoder.Device,
	display *tm1637.Device,
	monthIdx *int,
	dayIdx *int,
	dss *dateSetState) {

	display.DisplayClock(uint8(months[*monthIdx]), uint8(days[*dayIdx]), false)
	for {
		delta := <-enc.Dir
		if !dss.monthIsSet {
			*monthIdx = capIdx(*monthIdx+int(delta), len(months), true)
			*dayIdx = capIdx(*dayIdx, daysInMonth[*monthIdx]-1, false)
		} else if !dss.dayIsSet {
			*dayIdx = capIdx(*dayIdx+int(delta), daysInMonth[*monthIdx], true)
		}
		display.DisplayClock(uint8(months[*monthIdx]), uint8(days[*dayIdx]), false)
	}
}

// setDateState updates the date state based on the rotary encoder input.
//
// Parameters:
// - enc: a pointer to the rotary encoder device.
// - dss: a pointer to the date set state.
func setDateState(enc *rotaryencoder.Device,
	dss *dateSetState) {
	var curStateIndex int
	for {
		if <-enc.Switch {
			curStateIndex++
			i := curStateIndex % len(dateSetStates) // cicrular around the array
			*dss = dateSetStates[i]
		}
	}
}

func main() {
	emptyChan := make(chan bool)

	monthIdx := 0
	dayIdx := 0
	dss := dateSetStates[0]
	dateEnc := rotaryencoder.New(machine.GP7, machine.GP6, machine.GP8)
	dateEnc.Configure()

	dateDisplay := tm1637.New(machine.GP10, machine.GP11, 7) // clk, dio, brightness
	dateDisplay.Configure()
	dateDisplay.ClearDisplay()

	go setDateState(&dateEnc, &dss)
	go setDate(&dateEnc, &dateDisplay, &monthIdx, &dayIdx, &dss)

	// Start the main loop.
	// The blocking operation of reading from an empty channel allows the goroutines to run
	for {
		<-emptyChan
	}
}
