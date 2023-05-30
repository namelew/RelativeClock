package main

import "github.com/namelew/RelativeClock/internal/database"

func main() {
	routine := database.New()

	routine.Run()
}
