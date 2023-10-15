package setyear

import (
	"github.com/pavelanni/bttf/sound"
	"github.com/pavelanni/tinygo-drivers/rotaryencoder"
	"github.com/pavelanni/tinygo-drivers/tm1637"
)

// SetYear sets the year value using a rotary encoder and displays it on a TM1637 device.
//
// It takes the following parameters:
// - enc: A pointer to a rotaryencoder.Device struct representing the rotary encoder.
// - display: A pointer to a tm1637.Device struct representing the TM1637 device.
// - year: A pointer to an int16 variable representing the year to be displayed.
// - yearIsSet: A pointer to a bool variable indicating whether the year has been set.
//
// The function does not return any value.
func SetYear(enc *rotaryencoder.Device, display *tm1637.Device, year *int16, yearIsSet *bool) {
	display.DisplayNumber(*year)
	for {
		delta := <-enc.Dir
		if !*yearIsSet {
			*year += int16(delta)
			display.DisplayNumber(*year)
		}
	}
}

// SetYearBoolean sets the value of the yearIsSet boolean based on the state of the rotary encoder switch.
//
// It takes the following parameters:
// - enc: A pointer to a rotaryencoder.Device struct representing the rotary encoder.
// - yearIsSet: A pointer to a bool variable indicating whether the year has been set.
//
// The function does not return any value.
func SetYearBoolean(enc *rotaryencoder.Device, yearIsSet *bool) {
	for {
		if <-enc.Switch {
			*yearIsSet = !*yearIsSet
			println("year is set: ", *yearIsSet)
			if *yearIsSet {
				go sound.Player.Play(sound.Effects["set"])
			}
		}
	}
}
