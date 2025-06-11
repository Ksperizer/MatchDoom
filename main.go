package main

import (
	"MatchDoom/back"
	"MatchDoom/data"
)

func main() {
	data.InitDB()
	defer data.DB.Close()
	
	back.Server()
}
