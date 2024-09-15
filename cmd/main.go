package main

import (
	"log"

	"github.com/r-mol/Test_Service/internal/app"
)

func main() {
	if err := app.GetApp().Execute(); err != nil {
		log.Fatal(err)
	}
}
