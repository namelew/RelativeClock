package main

import (
	"github.com/joho/godotenv"
	"github.com/namelew/RelativeClock/internal/client"
)

func main() {
	godotenv.Load()

	pd := client.New()

	pd.Run()
}
