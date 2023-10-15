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

	rtcScl = machine.GP21
	rtcSda = machine.GP20
)

const (
	initialPresent   = "1985-10-26T01:22:00Z"
	initialLast      = "1985-10-26T01:20:00Z"
	savedPresentLast = "1985-10-26T01:20:00Z 1985-10-26T01:22:00Z"
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

var tPresent, tLast time.Time

var destChan = make(chan string)

func configureUart() {
	uart.Configure(machine.UARTConfig{
		BaudRate: 115200,
		TX:       tx,
		RX:       rx})
}

func NewDisplay(yearClk, yearDt, dateClk, dateDt, timeClk, timeDt machine.Pin) Display {
	const brightness uint8 = 7

	d := Display{
		Year: tm1637.New(yearClk, yearDt, brightness),
		Date: tm1637.New(dateClk, dateDt, brightness),
		Time: tm1637.New(timeClk, timeDt, brightness),
	}
	d.Year.Configure()
	d.Date.Configure()
	d.Time.Configure()
	d.Year.ClearDisplay()
	d.Date.ClearDisplay()
	d.Time.ClearDisplay()
	return d

}

func (d Display) Show(t time.Time) {
	year := int16(t.Year())
	monthIdx := int(t.Month()) - 1
	dayIdx := t.Day() - 1
	hour := uint8(t.Hour())
	minute := uint8(t.Minute())

	d.Year.DisplayNumber(year)
	d.Date.DisplayClock(uint8(setdate.Months[monthIdx]), uint8(setdate.Days[dayIdx]), false)
	d.Time.DisplayClock(hour, minute, true)
}

func readUart() {
	for {
		data := make([]byte, 1)
		if uart.Buffered() > 0 {
			discard, _ := uart.ReadByte()
			println("discarded: ", discard)
		}

		for {
			time.Sleep(10 * time.Millisecond)
			if uart.Buffered() > 0 {
				inByte, _ := uart.ReadByte()
				if inByte != byte('\n') {
					data = append(data, inByte)
					continue
				} else {
					break
				}
			}
		}
		println("read from UART: ", string(data))
		select {
		case destChan <- string(data):
		default:
		}
	}
}

func showPresent(d Display) {
	for {
		tPresent = time.Now()
		d.Show(tPresent)
		// save tPresent and tLast to flash
		newPresentLast := tPresent.Format(time.RFC3339) + " " + tLast.Format(time.RFC3339)
		writeFlash([]byte(newPresentLast))
		time.Sleep(1 * time.Second)
	}
}

func readFlash(data []byte) {
	println("reading flash...")
	_, err := machine.Flash.ReadAt(data, 0)
	if err != nil {
		log.Fatal(err)
	}
	println("read from flash:", string(data))
}

func writeFlash(data []byte) {
	//println("erasing flash...")
	needed := int64(len(savedPresentLast) / int(machine.Flash.EraseBlockSize()))
	if needed == 0 {
		needed = 1
	}
	err := machine.Flash.EraseBlocks(0, needed)
	if err != nil {
		log.Fatal(err)
	}

	//println("writing to flash: ", string(data))
	_, err = machine.Flash.WriteAt(data, 0)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var err error
	configureUart()

	time.Sleep(2 * time.Second)
	buffer := make([]byte, len(savedPresentLast))
	readFlash(buffer)
	tPresent, err = time.Parse(time.RFC3339, string(buffer[:20]))
	if err != nil {
		println("no present time in flash, setting tPresent to ", initialPresent)
		tPresent, err = time.Parse(time.RFC3339, initialPresent)
		if err != nil {
			log.Fatal(err)
		}
	}
	tLast, err = time.Parse(time.RFC3339, string(buffer[21:])) // 21 because of the space
	if err != nil {
		println("no last departed time in flash, setting tLast to ", initialLast)
		tLast, err = time.Parse(time.RFC3339, initialLast)
		if err != nil {
			log.Fatal(err)
		}
	}
	// update Now()
	timeOfMeasurement := time.Now()
	offset := tPresent.Sub(timeOfMeasurement)
	runtime.AdjustTimeOffset(int64(offset))

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

	dPresent.Show(tPresent)
	dLast.Show(tLast)

	go readUart()
	go showPresent(dPresent)
	for {
		destRFC3339 := <-destChan
		tDest, err := time.Parse(time.RFC3339, destRFC3339[1:])
		if err != nil {
			log.Println(err)
		}
		tLast = tPresent // last departed becomes the previous current
		tPresent = tDest // the new current we get from the destination
		// update Now()
		timeOfMeasurement := time.Now()
		offset := tPresent.Sub(timeOfMeasurement)
		runtime.AdjustTimeOffset(int64(offset))
		println("tPresent: ", tPresent.Format(time.RFC3339))
		println("time.Now() : ", time.Now().Format(time.RFC3339))
		// update displays
		dPresent.Show(tPresent)
		dLast.Show(tLast)
		newPresentLast := tPresent.Format(time.RFC3339) + " " + tLast.Format(time.RFC3339)
		writeFlash([]byte(newPresentLast))
	}
}
