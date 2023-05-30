package main

import "github.com/namelew/RelativeClock/internal/fees"

func main() {
	routine := fees.New()

	routine.Run()
}
