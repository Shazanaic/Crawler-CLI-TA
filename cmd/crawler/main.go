package main

import (
	"Crawler-CLI-TA/base/config"
	"context"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		fmt.Println("Error parsing config:", err)
		os.Exit(1)
	}

	fmt.Println("Config:", cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
}
