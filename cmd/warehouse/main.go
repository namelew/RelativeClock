package main

import (
	"github.com/joho/godotenv"
	"github.com/namelew/RelativeClock/internal/warehouse"
)

func main() {
	godotenv.Load()

	pd := warehouse.New()

	pd.Run()
}
