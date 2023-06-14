package main

import (
	"github.com/Zephyr75/gutter/test"
	"github.com/Zephyr75/gutter/app"
)

func main() {

  app := app.App {
    Name: "Gutter",
    Width: 800,
    Height: 600,
  }

  app.Run(test.MainWindow)

}
