package main

import (
	"github.com/joho/godotenv"
	"github.com/namelew/RelativeClock/internal/database"
)

func main() {
	godotenv.Load()

	routine := database.New()

	routine.Run()
}
