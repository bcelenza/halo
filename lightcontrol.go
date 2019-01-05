package main

import (
	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
	"log"
	"time"
)

type LightControl struct {
	location *Location
	lighting *Lighting
	plm *plm.PLM
}

func NewLightControl(location *Location, lighting *Lighting, plm *plm.PLM) *LightControl {
	return &LightControl{
		location:location,
		lighting:lighting,
		plm:plm,
	}
}

// Start the lighting control loop with a given duration. The done channel is currently not implemented.
func (lc *LightControl) Start(interval time.Duration, doneChan chan bool) {
	for {
		log.Println("Updating lighting brightness")

		// get the latest outdoor scene
		now := time.Now()
		outdoorScene, err := lc.location.GetOutdoorScene(now)
		if err != nil {
			log.Printf("Unable to get OutdoorScene for this interval: %s\n", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		for _, light := range lc.lighting.Lights {
			if light.Type != Dimmer.String() {
				log.Printf("Light '%s' is not a dimmer, skipping\n", light.Name)
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
		log.Printf("Could not get desired brightness for light '%s': %s\n", light.Name, err)
		return
	}

	log.Printf("Setting desired brightness for light '%s' to %d%%\n", light.Name, int(desiredBrightnessPercent*100))
	for _, address := range light.InsteonOptions.Addresses {
		addr := insteon.Address{}
		err := addr.UnmarshalText([]byte(address))
		if err != nil {
			log.Printf("Could not parse insteon address for light '%s', address %s: %s\n", light.Name, address, err)
			return
		}

		device, err := modem.Network.Connect(addr)
		if err != nil {
			log.Printf("Could not connect to device at address %s: %s\n", address, err)
			return
		}

		if dimmer, ok := device.(insteon.Dimmer); ok {
			// Determine the onLevel
			onLevel := int(desiredBrightnessPercent * 255)
			log.Printf("Calculated onLevel for device at address %s is %d\n", address, onLevel)

			// Set the default onLevel
			err = dimmer.SetDefaultOnLevel(onLevel)
			if err != nil {
				log.Printf("Could not set default OnLevel for device at address %s: %s\n", address, err)
			}

			// If the light is turned on, optionally adjust it's brightness
			currentOnLevel, err := dimmer.Status()
			if err != nil {
				log.Printf("Could not determing current OnLevel for device at address %s: %s\n", address, err)
				return
			}

			log.Printf("Current onlevel for address %s is %d\n", address, currentOnLevel)
			if currentOnLevel > 0 && ((!scene.AfterMeridian && currentOnLevel < onLevel) || (scene.AfterMeridian && currentOnLevel > onLevel)) {
				log.Printf("Adjusting current OnLevel of device at address %s to %d\n", address, onLevel)
				err = dimmer.OnAtRamp(onLevel, 60)
				if err != nil {
					log.Printf("Could not adjust OnLevel of device at address %s: %s\n", address, err)
					return
				}
			}
		} else {
			log.Printf("Device at address %s is not a dimmer", address)
			return
		}
	}
}
