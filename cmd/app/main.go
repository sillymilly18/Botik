package main

import (
	"context"

	"itmostar/internal/app"
	"itmostar/internal/config"
)

func main() {
	config.MustSetup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.New().Run(ctx); err != nil {
		panic(err)
	}
}
