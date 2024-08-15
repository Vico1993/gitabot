package main

import (
	"fmt"
	"log"

	service "github.com/Vico1993/gitabot/internal/services"
	"github.com/subosito/gotenv"
)

func main() {
	// load .env file if any otherwise use env set
	_ = gotenv.Load()

	// Load service
	err := service.Init()
	if err != nil {
		fmt.Println("Issue initialtion services")
		log.Fatalln(err)
	}

	fmt.Println("Hello Victor")
}
