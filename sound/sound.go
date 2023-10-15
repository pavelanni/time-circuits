package sound

import (
	"machine"

	"github.com/pavelanni/tinygo-drivers/dfplayermini"
)

var Player *dfplayermini.Device

var (
	uart1   = machine.UART1
	tx1     = machine.GP8
	rx1     = machine.GP9
	Effects = map[string]uint16{
		"poweron": 1,
		"set":     2,
		"jump":    3,
		"bell":    4,
		"click":   5,
	}
)

func ConfigurePlayer() {
	Player = dfplayermini.New(uart1, tx1, rx1)
	Player.Configure()
}
