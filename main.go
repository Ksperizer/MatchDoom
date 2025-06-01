package main 

import (
	"log"
	"net/http"
	"MatchDoom/data"
	"MatchDoom/rooter"
)

func main() {
	data.InitDB() 
	r := rooter.NewRouter()

	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}