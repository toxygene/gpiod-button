package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/toxygene/gpiod-button/device"
	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
	"golang.org/x/sync/errgroup"
)

func main() {
	actions := make(chan device.Action)

	g, ctx := errgroup.WithContext(context.Background())
	childCtx, cancel := context.WithCancel(ctx)

	g.Go(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		select {
		case <-c:
			cancel()
		case <-childCtx.Done():
		}

		return nil
	})

	g.Go(func() error {
		defer close(actions)

		chip, err := gpiod.NewChip("gpiochip0")
		if err != nil {
			return fmt.Errorf("create gpiod.chip: %w", err)
		}

		defer chip.Close()

		b := device.NewButton(chip, rpi.GPIO23)

		return b.Run(childCtx, actions)
	})

	g.Go(func() error {
		for action := range actions {
			switch action {
			case device.Press:
				println("Press")
			case device.Release:
				println("Release")
			}
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		println(err.Error())
		os.Exit(1)
	}
}