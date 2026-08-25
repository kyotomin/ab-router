package main

import (
	"fmt"
	"log"

	"github.com/kyotomin/ab-router/internal/app"
	"github.com/kyotomin/ab-router/internal/config"
)

func main() {
	cfg := config.MustLoad()

	app := app.New(cfg)

	fmt.Println("launching app...")
	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
