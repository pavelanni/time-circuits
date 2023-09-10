package main

import (
	"machine"

	"github.com/pavelanni/tinygo-drivers/rotaryencoder"
	"github.com/pavelanni/tinygo-drivers/tm1637"
)

const (
	yearDisplayClk = machine.GP0
	yearDisplayDt  = machine.GP1
	yearEncClk     = machine.GP4
	yearEncDt      = machine.GP5
	yearEncSwitch  = machine.GP6
)

// setYear sets the year value using a rotary encoder and displays it on a TM1637 device.
//
// It takes the following parameters:
// - enc: A pointer to a rotaryencoder.Device struct representing the rotary encoder.
// - display: A pointer to a tm1637.Device struct representing the TM1637 device.
// - year: A pointer to an int16 variable representing the year to be displayed.
// - yearIsSet: A pointer to a bool variable indicating whether the year has been set.
//
// The function does not return any value.
func setYear(enc *rotaryencoder.Device, display *tm1637.Device, year *int16, yearIsSet *bool) {
	display.DisplayNumber(*year)
	for {
		delta := <-enc.Dir
		if !*yearIsSet {
			*year += int16(delta)
			display.DisplayNumber(*year)
		}
	}
}

// setYearBoolean sets the value of the yearIsSet boolean based on the state of the rotary encoder switch.
//
// It takes the following parameters:
// - enc: A pointer to a rotaryencoder.Device struct representing the rotary encoder.
// - yearIsSet: A pointer to a bool variable indicating whether the year has been set.
//
// The function does not return any value.
func setYearBoolean(enc *rotaryencoder.Device, yearIsSet *bool) {
	for {
		if <-enc.Switch {
			*yearIsSet = !*yearIsSet
			println("year is set: ", *yearIsSet)
		}
	}
}

func main() {
	var year int16 = 2023
	var yearIsSet bool = true
	emptyChan := make(chan bool)

	yearEnc := rotaryencoder.New(yearEncClk, yearEncDt, yearEncSwitch)
	yearEnc.Configure()

	yearDisplay := tm1637.New(yearDisplayClk, yearDisplayDt, 7) // clk, dio, brightness
	yearDisplay.Configure()
	yearDisplay.ClearDisplay()

	go setYearBoolean(&yearEnc, &yearIsSet)
	go setYear(&yearEnc, &yearDisplay, &year, &yearIsSet)

	for {
		<-emptyChan
	}
}
