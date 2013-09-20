package main

import (
	"fmt"
	"github.com/larzconwell/moln/config"
	"log"
	"os"
)

func main() {
	env := "development"
	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	conf, err := config.Read("config/environment.json", "config/"+env+".json")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(conf)
}
