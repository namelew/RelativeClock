package main

import (
	"github.com/joho/godotenv"
	"github.com/namelew/RelativeClock/internal/server"
)

func main() {
	godotenv.Load()

	pd := server.New()

	pd.Run()
}
