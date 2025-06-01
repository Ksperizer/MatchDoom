package main

import (
	"MatchDoom/data"
	"MatchDoom/rooter"
	"log"
	"net/http"
	"MatchDoom/back"
)

func main() {
	data.InitDB() 
	r := rooter.NewRouter()

	go back.StartTCPServer() 


	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}



