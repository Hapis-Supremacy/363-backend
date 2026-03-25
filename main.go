package main

import (
	"363project/controller"
	"363project/initializer"
	"363project/middleware"
	"log"
	"net/http"
)

func init() {
	initializer.LoadEnvVar()   // Baca .env
	initializer.ConnecttoDB()  // Konek MySQL
	initializer.SyncDatabase() // Bikin tabel otomatis
}

func main() {
	// Pasang rute /login yang dilindungi AuthMiddleware
	http.HandleFunc("/login", middleware.AuthMiddleware(controller.USSDHandler))

	log.Println("Server 363-Project jalan di http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
