package main

import (
	"fmt"
	"time"

	"github.com/oskoss/mi-casa/config"
)

func main() {
	configFile := config.YamlConfig{FileLocation: "config.yaml"}
	micasaConfig, err := configFile.GetAllFields()
	fmt.Println(err)
	err = micasaConfig.DysonHotCoolLink[0].Connect()
	fmt.Println(err)
	time.Sleep(30 * time.Second)
	temp, err := micasaConfig.DysonHotCoolLink[0].CurrentTemp()
	fmt.Println(err)
	fmt.Println(*temp)

}
