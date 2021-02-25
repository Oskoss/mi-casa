package main

import (
	"fmt"
	"log"
	"time"

	"github.com/oskoss/mi-casa/config"
)

func main() {
	configFile := config.YamlConfig{FileLocation: "config.yaml"}
	micasaConfig, err := configFile.GetAllFields()
	if err != nil {
		log.Fatal(err)
	}
	err = micasaConfig.DysonHotCoolLink[0].Connect()
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(30 * time.Second) // why?
	temp, err := micasaConfig.DysonHotCoolLink[0].CurrentTemp()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*temp)

}
