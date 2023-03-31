package device

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
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
	logger  *logrus.Entry
}

func NewButton(chip *gpiod.Chip, dataPint int, logger *logrus.Entry) *Button {
	return &Button{
		chip:    chip,
		dataPin: dataPint,
		logger:  logger,
	}
}

func (t *Button) Run(ctx context.Context, actions chan<- Action) error {
	t.logger.Info("button started")
	defer t.logger.Info("button finished")

	var line *gpiod.Line

	previousValue := 1

	handler := func(event gpiod.LineEvent) {
		t.logger.Info("button event handler started")
		defer t.logger.Info("button event handler finished")

		value, err := line.Value()
		if err != nil {
			t.logger.WithError(err).Error("read button line value failed")
			panic(err)
		}

		t.logger.WithField("line", line).WithField("previousValue", previousValue).WithField("value", value).Trace("read button line value")

		if value != previousValue {
			if value == 1 {
				actions <- Release
			} else {
				actions <- Press
			}

			previousValue = value
		}
	}

	line, err := t.chip.RequestLine(t.dataPin, gpiod.AsInput, gpiod.WithPullUp, gpiod.WithBothEdges, gpiod.WithEventHandler(handler))
	if err != nil {
		lineInfo, _ := t.chip.LineInfo(t.dataPin)

		t.logger.WithError(err).WithField("dataPin", t.dataPin).WithField("dataLineInfo", lineInfo).Error("request button line failed")
		return fmt.Errorf("request button line: %w", err)
	}

	defer line.Close()

	<-ctx.Done()

	return nil
}
