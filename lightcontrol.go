package main

import (
	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
	log "github.com/sirupsen/logrus"
	"time"
)

type LightControl struct {
	location *Location
	lighting *Lighting
	plm      *plm.PLM
}

func NewLightControl(location *Location, lighting *Lighting, plm *plm.PLM) *LightControl {
	return &LightControl{
		location: location,
		lighting: lighting,
		plm:      plm,
	}
}

// Start the lighting control loop with a given duration. The done channel is currently not implemented.
func (lc *LightControl) Start(interval time.Duration, doneChan chan bool) {
	for {
		log.Info("Updating lighting brightness")

		// get the latest outdoor scene
		now := time.Now()
		outdoorScene, err := lc.location.GetOutdoorScene(now)
		if err != nil {
			log.Error("Unable to get OutdoorScene for this interval: ", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		for _, light := range lc.lighting.Lights {
			if light.Type != Dimmer.String() {
				log.WithField("name", light.Name).Warn("Light is not a dimmer, skipping")
				continue
			}

			lc.setBrightness(light, outdoorScene)
			time.Sleep(2 * time.Second)
		}

		time.Sleep(interval)
	}
}

func (*LightControl) setBrightness(light Light, scene *OutdoorScene) {
	desiredBrightnessPercent, err := light.DesiredBrightnessPercent(scene)
	if err != nil {
		log.WithField("name", light.Name).Error("Could not get desired brightness for light: ", err)
		return
	}

	log.WithFields(log.Fields{
		"name":       light.Name,
		"brightness": int(desiredBrightnessPercent * 100),
	}).Info("Setting desired brightness for light")
	for _, address := range light.InsteonOptions.Addresses {
		addr := insteon.Address{}
		err := addr.UnmarshalText([]byte(address))
		if err != nil {
			log.WithFields(log.Fields{
				"name":    light.Name,
				"address": address,
			}).Error("Could not parse insteon address for light: ", err)
			return
		}

		device, err := modem.Network.Connect(addr)
		if err != nil {
			log.WithField("address", address).Error("Could not connect to device: ", err)
			return
		}

		if dimmer, ok := device.(insteon.Dimmer); ok {
			// Determine the onLevel
			onLevel := int(desiredBrightnessPercent * 255)
			log.WithField("address", address).Debug("Calculated OnLevel for device is ", onLevel)

			// Set the default onLevel
			err = dimmer.SetDefaultOnLevel(onLevel)
			if err != nil {
				log.WithField("address", address).Error("Could not set default OnLevel for device: ", err)
			}

			// If the light is turned on, optionally adjust it's brightness
			currentOnLevel, err := dimmer.Status()
			if err != nil {
				log.WithField("address", address).Error("Could not determine current OnLevel for device: ", err)
				return
			}

			log.WithField("address", address).Debug("Current OnLevel is ", currentOnLevel)
			if currentOnLevel > 0 && ((!scene.AfterMeridian && currentOnLevel < onLevel) || (scene.AfterMeridian && currentOnLevel > onLevel)) {
				log.WithField("address", address).Info("Adjusting current OnLevel to ", onLevel)
				err = dimmer.OnAtRamp(onLevel, 60)
				if err != nil {
					log.WithField("address", address).Error("Could not adjust OnLevel: ", err)
					return
				}
			}
		} else {
			log.WithField("address", address).Error("Device is not a dimmer")
			return
		}
	}
}
