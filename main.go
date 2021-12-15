package main

import (
	"github.com/satmaelstorm/filup/cmd"
	"log"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Println(err)
	}
}
