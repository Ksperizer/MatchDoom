package main

import (
	"MatchDoom/data"
	"MatchDoom/rooter"
	"log"
	"net/http"
	"MatchDoom/back"
)

func main() {
	data.InitDB() // init database 
	go back.StartTCPServer() 
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./template"))
	mux.Handle("/template/", http.StripPrefix("/template/", fs))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./template/html/accueil.html")
	})

	apiRouter := rooter.NewRouter()
	mux.Handle("/api/", apiRouter)

	// start the server
	port := "8080"
	log.Printf(" Serveur lancé : http://localhost:%s", port)
	log.Println("Ouvrez ce lien dans votre navigateur ")

	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatalf("Erreur serveur : %v", err)
	}
}



