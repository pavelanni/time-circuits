module setyear

go 1.20

//replace github.com/pavelanni/tinygo-drivers/tm1637 => /home/pavel/Projects/tinygo-drivers/tm1637

//replace github.com/pavelanni/tinygo-drivers/rotaryencoder => /home/pavel/Projects/tinygo-drivers/rotaryencoder

require (
	github.com/pavelanni/tinygo-drivers/rotaryencoder v0.0.0-20230813234026-d978e2a95923
	github.com/pavelanni/tinygo-drivers/tm1637 v0.0.0-20230813002143-4c86b787d857
)
