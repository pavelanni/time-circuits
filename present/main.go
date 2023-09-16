package main

import (
	"log"
	"machine"
	"time"

	"github.com/pavelanni/bttf/setdate"
	"github.com/pavelanni/tinygo-drivers/tm1637"
)

const (
	yearPresentDisplayClk = machine.GP0
	yearPresentDisplayDt  = machine.GP1
	datePresentDisplayClk = machine.GP2
	datePresentDisplayDt  = machine.GP3
	timePresentDisplayClk = machine.GP16
	timePresentDisplayDt  = machine.GP17
	yearLastDisplayClk    = machine.GP4
	yearLastDisplayDt     = machine.GP5
	dateLastDisplayClk    = machine.GP6
	dateLastDisplayDt     = machine.GP7
	timeLastDisplayClk    = machine.GP18
	timeLastDisplayDt     = machine.GP19
)

func main() {
	emptyChan := make(chan bool)

	initialPresent := "1985-10-26T01:22:00Z"
	initialLast := "1985-10-26T01:20:00Z"
	tPresent, err := time.Parse(time.RFC3339, initialPresent)
	if err != nil {
		log.Fatal(err)
	}
	tLast, err := time.Parse(time.RFC3339, initialLast)
	if err != nil {
		log.Fatal(err)
	}

	yearPresentDisplay := tm1637.New(yearPresentDisplayClk, yearPresentDisplayDt, 0)
	yearPresentDisplay.Configure()
	yearPresentDisplay.ClearDisplay()
	yearLastDisplay := tm1637.New(yearLastDisplayClk, yearLastDisplayDt, 0)
	yearLastDisplay.Configure()
	yearLastDisplay.ClearDisplay()

	yearPresent := int16(tPresent.Year())
	yearLast := int16(tLast.Year())
	monthPresentIdx := int(tPresent.Month()) - 1
	dayPresentIdx := tPresent.Day() - 1
	monthLastIdx := int(tLast.Month()) - 1
	dayLastIdx := tLast.Day() - 1
	hourPresent := uint8(tPresent.Hour())
	minutePresent := uint8(tPresent.Minute())
	hourLast := uint8(tLast.Hour())
	minuteLast := uint8(tLast.Minute())

	datePresentDisplay := tm1637.New(datePresentDisplayClk, datePresentDisplayDt, 0)
	datePresentDisplay.Configure()
	datePresentDisplay.ClearDisplay()
	dateLastDisplay := tm1637.New(dateLastDisplayClk, dateLastDisplayDt, 0)
	dateLastDisplay.Configure()
	dateLastDisplay.ClearDisplay()

	timePresentDisplay := tm1637.New(timePresentDisplayClk, timePresentDisplayDt, 0) // clk, dio, brightness
	timePresentDisplay.Configure()
	timePresentDisplay.ClearDisplay()
	timeLastDisplay := tm1637.New(timeLastDisplayClk, timeLastDisplayDt, 0) // clk, dio, brightness
	timeLastDisplay.Configure()
	timeLastDisplay.ClearDisplay()

	yearPresentDisplay.DisplayNumber(yearPresent)
	datePresentDisplay.DisplayClock(uint8(setdate.Months[monthPresentIdx]), uint8(setdate.Days[dayPresentIdx]), false)
	timePresentDisplay.DisplayClock(hourPresent, minutePresent, true)
	go yearPresentDisplay.FadeIn(4 * time.Second)
	go datePresentDisplay.FadeIn(4 * time.Second)
	go timePresentDisplay.FadeIn(4 * time.Second)
	yearLastDisplay.DisplayNumber(yearLast)
	dateLastDisplay.DisplayClock(uint8(setdate.Months[monthLastIdx]), uint8(setdate.Days[dayLastIdx]), false)
	timeLastDisplay.DisplayClock(hourLast, minuteLast, true)
	go yearLastDisplay.FadeIn(4 * time.Second)
	go dateLastDisplay.FadeIn(4 * time.Second)
	go timeLastDisplay.FadeIn(4 * time.Second)

	for {
		<-emptyChan
	}
}
