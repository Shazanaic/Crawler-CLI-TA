package main

import (
	"fmt"
	"ints-test-assign/base/config"
	"os"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		fmt.Println("Error parsing config:", err)
		os.Exit(1)
	}

	fmt.Println("Config:", cfg)

}
