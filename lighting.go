package main

import (
	"errors"
)

type Protocol int

const (
	Insteon Protocol = iota
)

func (protocol Protocol) String() string {
	return [...]string{
		"insteon",
	}[protocol]
}

type LightType int

const (
	Standard LightType = iota
	Dimmer
)

func (lightType LightType) String() string {
	return [...]string{
		"standard",
		"dimmer",
	}[lightType]

}

type InsteonOptions struct {
	Addresses []string
}

type DimmerOptions struct {
	MinBrightnessPercent float64 `yaml:"min_brightness_percent"`
	MaxBrightnessPercent float64 `yaml:"max_brightness_percent"`
}

type Light struct {
	Name           string
	Type           string
	Protocol       string
	InsteonOptions InsteonOptions `yaml:"insteon_options"`
	DimmerOptions  DimmerOptions  `yaml:"dimmer_options"`
}

type Lighting struct {
	Lights []Light
}

func (light *Light) DesiredBrightnessPercent(outdoorScene *OutdoorScene) (float64, error) {
	if light.Type != Dimmer.String() {
		return 0, errors.New("light is not a dimmer")
	}

	// Determine the desired brightness between its min and max values
	desiredBrightness := (outdoorScene.BrightnessCoefficient-0)*
		(light.DimmerOptions.MaxBrightnessPercent-light.DimmerOptions.MinBrightnessPercent)/
		(outdoorScene.LightWindow-0) +
		light.DimmerOptions.MinBrightnessPercent

	return desiredBrightness / 100, nil
}
