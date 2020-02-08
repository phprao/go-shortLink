package main

import (
	"fmt"
	"shortLink/app"
)

func main() {
	a := app.App{}
	a.Initialize()
	fmt.Println("The Server is listening on 8080...")
	a.Run(":8080")
}