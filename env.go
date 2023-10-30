package main

import (
	"time"

	"github.com/caarlos0/env/v6"
)

type Env struct {
	MeasurementID string        `env:"MEASUREMENT_ID,required"`
	APISecret     string        `env:"API_SECRET,required"`
	TickInterval  time.Duration `env:"TICK_INTERVAL,required" envDefault:"10s"`
}

func (e *Env) Parse() error {
	if err := env.Parse(e); err != nil {
		return err
	}
	return nil
}
