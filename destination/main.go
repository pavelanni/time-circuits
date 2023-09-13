package main

import (
	"machine"
	"time"

	"github.com/pavelanni/bttf/setdate"
	"github.com/pavelanni/bttf/settime"
	"github.com/pavelanni/bttf/setyear"
	"github.com/pavelanni/tinygo-drivers/rotaryencoder"
	"github.com/pavelanni/tinygo-drivers/tm1637"
)

const (
	yearDisplayClk = machine.GP0
	yearDisplayDt  = machine.GP1
	yearEncClk     = machine.GP4
	yearEncDt      = machine.GP5
	yearEncSwitch  = machine.GP6
	dateDisplayClk = machine.GP2
	dateDisplayDt  = machine.GP3
	dateEncClk     = machine.GP21
	dateEncDt      = machine.GP22
	dateEncSwitch  = machine.GP26
	timeDisplayClk = machine.GP16
	timeDisplayDt  = machine.GP17
	timeEncClk     = machine.GP18
	timeEncDt      = machine.GP19
	timeEncSwitch  = machine.GP20
)

func main() {
	var year int16 = 2023
	var yearIsSet bool = true
	emptyChan := make(chan bool)

	yearEnc := rotaryencoder.New(yearEncClk, yearEncDt, yearEncSwitch)
	yearEnc.Configure()

	yearDisplay := tm1637.New(yearDisplayClk, yearDisplayDt, 0) // clk, dio, brightness
	yearDisplay.Configure()
	yearDisplay.ClearDisplay()

	monthIdx := 0
	dayIdx := 0
	dss := setdate.DateSetStates[0]
	dateEnc := rotaryencoder.New(dateEncClk, dateEncDt, dateEncSwitch)
	dateEnc.Configure()

	dateDisplay := tm1637.New(dateDisplayClk, dateDisplayDt, 0) // clk, dio, brightness
	dateDisplay.Configure()
	dateDisplay.ClearDisplay()

	hour := uint8(0)
	minute := uint8(0)
	tss := settime.TimeSetStates[0]
	timeEnc := rotaryencoder.New(timeEncClk, timeEncDt, timeEncSwitch)
	timeEnc.Configure()

	timeDisplay := tm1637.New(timeDisplayClk, timeDisplayDt, 0) // clk, dio, brightness
	timeDisplay.Configure()
	timeDisplay.ClearDisplay()

	yearDisplay.DisplayNumber(year)
	dateDisplay.DisplayClock(uint8(setdate.Months[monthIdx]), uint8(setdate.Days[dayIdx]), false)
	timeDisplay.DisplayClock(hour, minute, true)
	go yearDisplay.FadeIn(4 * time.Second)
	go dateDisplay.FadeIn(4 * time.Second)
	go timeDisplay.FadeIn(4 * time.Second)
	go setyear.SetYearBoolean(&yearEnc, &yearIsSet)
	go setyear.SetYear(&yearEnc, &yearDisplay, &year, &yearIsSet)
	go setdate.SetDateState(&dateEnc, &dss)
	go setdate.SetDate(&dateEnc, &dateDisplay, &monthIdx, &dayIdx, &dss)
	go settime.SetTimeState(&timeEnc, &tss)
	go settime.SetTime(&timeEnc, &timeDisplay, &hour, &minute, &tss)

	for {
		<-emptyChan
	}
}
