package main

import (
	"log"
	"machine"
	"time"

	"github.com/pavelanni/bttf/setdate"
	"github.com/pavelanni/bttf/settime"
	"github.com/pavelanni/bttf/setyear"
	"github.com/pavelanni/tinygo-drivers/rotaryencoder"
	"github.com/pavelanni/tinygo-drivers/tm1637"
)

const (
	yearDisplayClk = machine.GP2
	yearDisplayDt  = machine.GP3
	yearEncClk     = machine.GP4
	yearEncDt      = machine.GP5
	yearEncSwitch  = machine.GP6
	dateDisplayClk = machine.GP7
	dateDisplayDt  = machine.GP8
	dateEncClk     = machine.GP9
	dateEncDt      = machine.GP10
	dateEncSwitch  = machine.GP11
	timeDisplayClk = machine.GP12
	timeDisplayDt  = machine.GP13
	timeEncClk     = machine.GP14
	timeEncDt      = machine.GP15
	timeEncSwitch  = machine.GP16
)

const (
	initialDest = "1985-10-26T01:22:00Z"
)

var (
	uart      = machine.UART0
	tx        = machine.UART0_TX_PIN
	rx        = machine.UART0_RX_PIN
	buttonPin = machine.GP18
	bChan     = make(chan bool)
)

func configureUart() {
	uart.Configure(machine.UARTConfig{
		BaudRate: 115200,
		TX:       tx,
		RX:       rx})
}

func configureButton(p machine.Pin) {
	p.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	p.SetInterrupt(machine.PinFalling, buttonHandler)
}

func buttonHandler(p machine.Pin) {
	if !p.Get() {
		select {
		case bChan <- true:
		default:
		}
	}
}

func main() {
	configureUart()
	configureButton(buttonPin)
	// Set up year display and encoder
	var yearIsSet bool = true
	yearEnc := rotaryencoder.New(yearEncClk, yearEncDt, yearEncSwitch)
	yearEnc.Configure()
	yearDisplay := tm1637.New(yearDisplayClk, yearDisplayDt, 0) // clk, dio, brightness
	yearDisplay.Configure()
	yearDisplay.ClearDisplay()

	// Set up date display and encoder
	dss := setdate.DateSetStates[0]
	dateEnc := rotaryencoder.New(dateEncClk, dateEncDt, dateEncSwitch)
	dateEnc.Configure()
	dateDisplay := tm1637.New(dateDisplayClk, dateDisplayDt, 0) // clk, dio, brightness
	dateDisplay.Configure()
	dateDisplay.ClearDisplay()

	// Set up time display and encoder
	tss := settime.TimeSetStates[0]
	timeEnc := rotaryencoder.New(timeEncClk, timeEncDt, timeEncSwitch)
	timeEnc.Configure()
	timeDisplay := tm1637.New(timeDisplayClk, timeDisplayDt, 0) // clk, dio, brightness
	timeDisplay.Configure()
	timeDisplay.ClearDisplay()

	// Read time from the RFC3339 string
	timeDest, err := time.Parse(time.RFC3339, initialDest)
	if err != nil {
		log.Fatal(err)
	}
	year := int16(timeDest.Year())
	monthIdx := int(timeDest.Month()) - 1
	dayIdx := timeDest.Day() - 1
	hour := uint8(timeDest.Hour())
	minute := uint8(timeDest.Minute())

	// Display initial target date and time
	yearDisplay.DisplayNumber(year)
	dateDisplay.DisplayClock(uint8(monthIdx), uint8(dayIdx), false)
	timeDisplay.DisplayClock(hour, minute, true)
	go yearDisplay.FadeIn(4*time.Second, 7)
	go dateDisplay.FadeIn(4*time.Second, 7)
	go timeDisplay.FadeIn(4*time.Second, 7)
	go setyear.SetYearBoolean(&yearEnc, &yearIsSet)
	go setyear.SetYear(&yearEnc, &yearDisplay, &year, &yearIsSet)
	go setdate.SetDateState(&dateEnc, &dss)
	go setdate.SetDate(&dateEnc, &dateDisplay, &monthIdx, &dayIdx, &dss)
	go settime.SetTimeState(&timeEnc, &tss)
	go settime.SetTime(&timeEnc, &timeDisplay, &hour, &minute, &tss)

	for {
		if <-bChan {
			destDate := time.Date(int(year), time.Month(monthIdx+1), dayIdx+1, int(hour), int(minute), 0, 0, time.UTC)
			message := destDate.Format(time.RFC3339) + "\n"
			print(message)
			uart.Write([]byte(message))
		}
	}
}
