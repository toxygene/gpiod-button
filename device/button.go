package device

import (
	"context"
	"fmt"

	"github.com/warthog618/gpiod"
)

type Action string

var (
	Press   Action = "press"
	Release Action = "release"
)

type Button struct {
	chip    *gpiod.Chip
	dataPin int
}

func NewButton(chip *gpiod.Chip, dataPint int) *Button {
	return &Button{
		chip:    chip,
		dataPin: dataPint,
	}
}

func (b *Button) Run(ctx context.Context, actions chan<- Action) error {
	var line *gpiod.Line

	previousValue := 1

	handler := func(event gpiod.LineEvent) {
		value, err := line.Value()
		if err != nil {
			panic(err)
		}

		if value != previousValue {
			if value == 1 {
				actions <- Release
			} else {
				actions <- Press
			}

			previousValue = value
		}
	}

	line, err := b.chip.RequestLine(b.dataPin, gpiod.AsInput, gpiod.WithPullUp, gpiod.WithBothEdges, gpiod.WithEventHandler(handler))
	if err != nil {
		return fmt.Errorf("request data line: %w", err)
	}

	defer line.Close()

	<-ctx.Done()

	return nil
}
