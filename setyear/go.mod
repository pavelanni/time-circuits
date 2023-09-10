module setyear

go 1.20

//replace github.com/pavelanni/tinygo-drivers/tm1637 => /home/pavel/Projects/tinygo-drivers/tm1637

//replace github.com/pavelanni/tinygo-drivers/rotaryencoder => /home/pavel/Projects/tinygo-drivers/rotaryencoder

require (
	github.com/pavelanni/tinygo-drivers/rotaryencoder v0.0.0-20230910200355-5afcb9a9ef73
	github.com/pavelanni/tinygo-drivers/tm1637 v0.0.0-20230910200355-5afcb9a9ef73
)
