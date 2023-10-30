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
	initialDest      = "1985-10-26T01:22:00Z"
	initialPresent   = "1985-10-26T01:22:00Z"
	initialLast      = "1985-10-26T01:20:00Z"
	savedPresentLast = "1985-10-26T01:20:00Z 1985-10-26T01:22:00Z"
)

const (
	STX = byte(0x02) // ASCII Start transmission
	ETX = byte(0x03) // ASCII End transmission
)

var (
	uart     = machine.UART0
	tx       = machine.UART0_TX_PIN
	rx       = machine.UART0_RX_PIN
	buf      = make([]byte, 64) // why 64? just in case
	flashBuf = make([]byte, len(savedPresentLast))
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
		BaudRate: 9600,
		TX:       tx,
		RX:       rx})
	uart.SetFormat(8, 2, 0)
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

	//println("displaying year: ", year)
	d.Year.DisplayNumber(year)
	//println("displaying month, day: ", setdate.Months[monthIdx], setdate.Days[dayIdx])
	d.Date.DisplayClock(uint8(setdate.Months[monthIdx]), uint8(setdate.Days[dayIdx]), false)
	//println("displaying hour, minute: ", hour, minute)
	d.Time.DisplayClock(hour, minute, true)
}

func readUart() {
	println("reading UART...")
	for {
		time.Sleep(5 * time.Millisecond)
		if uart.Buffered() > 0 {
			inByte, err := uart.ReadByte()
			if err != nil {
				log.Println(err)
			}
			if inByte != STX { // waiting for the start of transmission
				continue
			}
		}
		i := 0
		for {
			time.Sleep(50 * time.Microsecond)
			if uart.Buffered() > 0 {
				//println("in the buffer: ", uart.Buffered(), " bytes")
				inByte, err := uart.ReadByte()
				if err != nil {
					log.Println(err)
				}
				if inByte != ETX {
					buf[i] = inByte
					i++
					//println("read byte: ", inByte)
					continue
				} else {
					break
				}
			}
		}
		println("read from UART: ", string(buf[:i]))
		select {
		case destChan <- string(buf[:i]):
			println("sent to destination channel: ", string(buf[:i]))
		default:
			println("could not send to destination channel")
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
		log.Println(err)
	}

	//println("writing to flash: ", string(data))
	_, err = machine.Flash.WriteAt(data, 0)
	if err != nil {
		log.Println(err)
	}
}

func cleanRFC3339(ts string) string {
	// remove the initial characters to leave only the string ending with Z that is the RFC3339 format
	// find the position of the first Z
	var index int
	for i, char := range ts {
		if char == 'Z' {
			index = i
			break
		}
	}
	if index < 19 {
		return ""
	}
	return ts[index-19 : index+1] // only the length of the RFC3339
}

func main() {
	var err error
	var tDest time.Time
	tDest, err = time.Parse(time.RFC3339, initialDest)
	if err != nil {
		log.Fatal(err) // wrong initialDest string
	}
	configureUart()

	time.Sleep(5 * time.Second)
	readFlash(flashBuf)
	tPresent, err = time.Parse(time.RFC3339, string(flashBuf[:20]))
	if err != nil {
		println("no present time in flash, setting tPresent to ", initialPresent)
		tPresent, err = time.Parse(time.RFC3339, initialPresent)
		if err != nil {
			log.Fatal(err)
		}
	}
	tLast, err = time.Parse(time.RFC3339, string(flashBuf[21:])) // 21 because of the space between the dates
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
		println("destRFC3339: ", destRFC3339)
		destRFC3339 = cleanRFC3339(destRFC3339)
		if destRFC3339 == "" {
			println("no destination time in UART, no change in tDest")
		} else {
			tDest, err = time.Parse(time.RFC3339, destRFC3339)
			if err != nil {
				log.Println(err)
			}
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
		println("written to flash: ", newPresentLast)
		time.Sleep(1 * time.Second)
	}
}
