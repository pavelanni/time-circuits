package main

import (
	"log"
	"machine"
	"runtime"
	"time"

	"github.com/pavelanni/bttf/setdate"
	"github.com/pavelanni/tinygo-drivers/tm1637"
)

const (
	yearPresentDisplayClk = machine.GP2
	yearPresentDisplayDt  = machine.GP3
	datePresentDisplayClk = machine.GP4
	datePresentDisplayDt  = machine.GP5
	timePresentDisplayClk = machine.GP6
	timePresentDisplayDt  = machine.GP7
	yearLastDisplayClk    = machine.GP8
	yearLastDisplayDt     = machine.GP9
	dateLastDisplayClk    = machine.GP10
	dateLastDisplayDt     = machine.GP11
	timeLastDisplayClk    = machine.GP12
	timeLastDisplayDt     = machine.GP13
)

const (
	initialPresent = "1985-10-26T01:22:00Z"
	initialLast    = "1985-10-26T01:20:00Z"
)

var (
	uart = machine.UART0
	tx   = machine.UART0_TX_PIN
	rx   = machine.UART0_RX_PIN
)

type Display struct {
	Year tm1637.Device
	Date tm1637.Device
	Time tm1637.Device
}

var tInitial time.Time

var destChan = make(chan string)

func configureUart() {
	uart.Configure(machine.UARTConfig{
		BaudRate: 115200,
		TX:       tx,
		RX:       rx})
}

func NewDisplay(yearClk, yearDt, dateClk, dateDt, timeClk, timeDt machine.Pin) Display {
	d := Display{
		Year: tm1637.New(yearClk, yearDt, 7),
		Date: tm1637.New(dateClk, dateDt, 7),
		Time: tm1637.New(timeClk, timeDt, 7),
	}
	d.Year.Configure()
	d.Date.Configure()
	d.Time.Configure()
	d.Year.ClearDisplay()
	d.Date.ClearDisplay()
	d.Time.ClearDisplay()
	return d

}

func (d Display) FadeIn(t time.Duration, brightness uint8) {
	go d.Year.FadeIn(t, brightness)
	go d.Date.FadeIn(t, brightness)
	go d.Time.FadeIn(t, brightness)
}

func (d Display) Brightness(b uint8) {
	d.Year.Brightness(b)
	d.Date.Brightness(b)
	d.Time.Brightness(b)
}

func (d Display) Show(t time.Time, brightness uint8) {
	year := int16(t.Year())
	monthIdx := int(t.Month()) - 1
	dayIdx := t.Day() - 1
	hour := uint8(t.Hour())
	minute := uint8(t.Minute())

	d.Year.DisplayNumber(year)
	d.Date.DisplayClock(uint8(setdate.Months[monthIdx]), uint8(setdate.Days[dayIdx]), false)
	d.Time.DisplayClock(hour, minute, true)
	d.Brightness(brightness)
	d.Brightness(brightness)
}

func readUart() {
	for {
		data := make([]byte, 1)
		if uart.Buffered() > 0 {
			discard, _ := uart.ReadByte()
			println("discarded: ", discard)
		}

		for {
			if uart.Buffered() > 0 {
				inByte, _ := uart.ReadByte()
				if inByte != byte('\n') {
					data = append(data, inByte)
					//runtime.Gosched()
					continue
				} else {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
		println(string(data))
		select {
		case destChan <- string(data):
		default:
		}
	}
}

func showPresent(d Display, brightness uint8) {
	for {
		d.Show(time.Now(), brightness)
		time.Sleep(5 * time.Second)
	}

}

func main() {
	configureUart()

	tPresent, err := time.Parse(time.RFC3339, initialPresent)
	if err != nil {
		log.Fatal(err)
	}
	tLast, err := time.Parse(time.RFC3339, initialLast)
	if err != nil {
		log.Fatal(err)
	}

	// Configure displays
	dPresent := NewDisplay(yearPresentDisplayClk,
		yearPresentDisplayDt,
		datePresentDisplayClk,
		datePresentDisplayDt,
		timePresentDisplayClk,
		timePresentDisplayDt)
	dLast := NewDisplay(yearLastDisplayClk,
		yearLastDisplayDt,
		dateLastDisplayClk,
		dateLastDisplayDt,
		timeLastDisplayClk,
		timeLastDisplayDt)

	dPresent.Show(tPresent, 7)
	dLast.Show(tLast, 7)
	//dPresent.FadeIn(4*time.Second, 7)
	//dLast.FadeIn(4*time.Second, 7)

	tInitial = tPresent // we use the global var tPresent here; for some reason nextTime doesn't work when I pass it as a param
	go readUart()
	go showPresent(dPresent, 7)
	for {
		destRFC3339 := <-destChan
		tDest, err := time.Parse(time.RFC3339, destRFC3339[1:]) // first byte is \x00 in the received string so we clip it
		if err != nil {
			log.Fatal(err)
		}
		tLast = tPresent
		tPresent = tDest
		timeOfMeasurement := time.Now()
		offset := tPresent.Sub(timeOfMeasurement)
		runtime.AdjustTimeOffset(int64(offset))
		dPresent.Show(time.Now(), 7)
		dLast.Show(tLast, 7)
	}
}
