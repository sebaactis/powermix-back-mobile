package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/sebaactis/powermix-back-mobile/internal/platform/config"
	"github.com/sebaactis/powermix-back-mobile/internal/platform/database"
	"github.com/sebaactis/powermix-back-mobile/internal/routes"
	"github.com/sebaactis/powermix-back-mobile/internal/validations"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error al cargar el archivo .env", err)
	}

	cfg := config.Load()

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatal("Error al conectar la base de datos", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatal("Error al migrar las entidades en la base de datos", err)
	}

	//validator := validations.NewValidator()

	r := routes.Router()

	log.Println("Servidor corriendo en el puerto :8080")
	http.ListenAndServe(":8080", r)
}
