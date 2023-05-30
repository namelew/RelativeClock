package main

import (
	"github.com/joho/godotenv"
	"github.com/namelew/RelativeClock/internal/fees"
)

func main() {
	godotenv.Load()
	routine := fees.New()

	routine.Run()
}
