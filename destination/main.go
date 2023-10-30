package main

import (
	"log"
	"machine"
	"time"

	"github.com/pavelanni/bttf/setdate"
	"github.com/pavelanni/bttf/settime"
	"github.com/pavelanni/bttf/setyear"
	"github.com/pavelanni/bttf/sound"
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
	dateDisplayDt  = machine.GP26
	dateEncClk     = machine.GP21
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

const (
	STX = byte(0x02) // ASCII Start transmission
	ETX = byte(0x03) // ASCII End transmission
)

var (
	uart             = machine.UART0
	tx               = machine.UART0_TX_PIN
	rx               = machine.UART0_RX_PIN
	buttonPin        = machine.GP18
	bChan            = make(chan bool)
	lastClick        time.Time
	debounceInterval time.Duration = 200 * time.Millisecond
)

func configureUart() {
	uart.Configure(machine.UARTConfig{
		BaudRate: 9600,
		TX:       tx,
		RX:       rx})
	uart.SetFormat(8, 2, 0)
}

func configureButton(p machine.Pin) {
	p.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	p.SetInterrupt(machine.PinFalling, buttonHandler)
}

func buttonHandler(p machine.Pin) {
	if !p.Get() {
		if time.Since(lastClick) > debounceInterval {
			lastClick = time.Now()
			select {
			case bChan <- true:
			default:
			}
		}
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
	needed := int64(len(initialDest) / int(machine.Flash.EraseBlockSize()))
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

func prepareString(s string) []byte {
	buf := make([]byte, len(s)+2)
	buf[0] = STX          // ASCII STX code -- start of transmission
	buf[len(buf)-1] = ETX // ASCII ETX code -- end of transmission
	copy(buf[1:len(buf)-1], s)
	return buf
}

func main() {
	configureUart()
	sound.ConfigurePlayer()
	configureButton(buttonPin)
	lastClick = time.Now()
	time.Sleep(2 * time.Second)
	go sound.Player.Play(sound.Effects["poweron"])

	buffer := make([]byte, len(initialDest))
	readFlash(buffer)
	tDest, err := time.Parse(time.RFC3339, string(buffer[:20]))
	if err != nil {
		println("no destination time in flash, setting tDest to ", initialDest)
		tDest, err = time.Parse(time.RFC3339, initialDest)
		if err != nil {
			log.Fatal(err)
		}
	}
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

	year := int16(tDest.Year())
	monthIdx := int(tDest.Month()) - 1
	dayIdx := tDest.Day() - 1
	hour := uint8(tDest.Hour())
	minute := uint8(tDest.Minute())

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
			//go sound.Player.Play(sound.Effects["jump"])
			destDate := time.Date(int(year), time.Month(monthIdx+1), dayIdx+1, int(hour), int(minute), 0, 0, time.UTC)
			message := prepareString(destDate.Format(time.RFC3339))
			println("sending to UART: ", string(message[1:len(message)-1]))
			/*
				for i := 0; i < len(message); i++ {
					uart.WriteByte(message[i])
					time.Sleep(5 * time.Microsecond)
				}
			*/
			n, err := uart.Write(message)
			if err != nil {
				log.Println(err)
			}
			println("sent to UART: ", n, " bytes")
			println("writing to flash: ", string(message[1:len(message)-1]))
			writeFlash(message[1 : len(message)-1])
		}
	}
}
