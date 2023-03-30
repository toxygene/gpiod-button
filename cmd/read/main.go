package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/toxygene/gpiod-button/device"
	"github.com/warthog618/gpiod"
	"golang.org/x/sync/errgroup"
)

func main() {
	buttonPinNumber := flag.Int("buttonPinNumber", 0, "GPIO pin number for the button")
	chipName := flag.String("chipName", "", "Chip name for the GPIO device of the rotary encoder and button")
	help := flag.Bool("help", false, "print help page")
	logging := flag.String("logging", "", "logging level")

	flag.Parse()

	if *help || *buttonPinNumber == 0 {
		flag.Usage()

		if *help {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	logger := logrus.New()

	if *logging != "" {
		logLevel, err := logrus.ParseLevel(*logging)
		if err != nil {
			println(fmt.Errorf("parse log level: %w", err).Error())
			os.Exit(1)
		}

		logger.SetLevel(logLevel)
	}

	chip, err := gpiod.NewChip(*chipName)
	if err != nil {
		panic(fmt.Errorf("create gpiod.chip: %w", err))
	}

	defer chip.Close()

	b := device.NewButton(chip, *buttonPinNumber, logrus.NewEntry(logger))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	actions := make(chan device.Action)

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		defer close(c)

		logger.Info("interrupt handler goroutine started")
		defer logger.Info("interrupt handler goroutine finished")

		select {
		case <-c:
			return errors.New("application interrupted")
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	g.Go(func() error {
		defer close(actions)

		logger.Info("button goroutine started")
		defer logger.Info("button goroutine finished")

		if err := b.Run(ctx, actions); err != nil {
			return fmt.Errorf("run button: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		logger.Info("action handler goroutine started")
		defer logger.Info("action handler goroutine finished")

		for action := range actions {
			logger.WithField("action", action).Trace("received action")

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
		println(fmt.Errorf("running application goroutines: %w", err).Error())
		os.Exit(1)
	}
}
