package main

import (
	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
	"github.com/tarm/serial"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"time"
)

var modem *plm.PLM

func main() {
	insteon.Log.Level(insteon.LevelDebug)
	log.Println("Hello")

	// parse home file
	home := Location{}
	homeFile, err := ioutil.ReadFile("config/home.yaml")
	if err != nil {
		log.Fatalf("Unable to open home configuration file: %s\n", err)
	}
	err = yaml.UnmarshalStrict(homeFile, &home)
	if err != nil {
		log.Fatalf("Unable to unmarshal home YAML: %s\n", err)
	}

	// parse lighting file
	lighting := Lighting{}

	lightingFile, err := ioutil.ReadFile("config/lighting.yaml")
	if err != nil {
		log.Fatalf("Unable to open lighting configuration file: %s\n", err)
	}

	err = yaml.UnmarshalStrict(lightingFile, &lighting)
	if err != nil {
		log.Fatalf("Unable to unmarshal lighting YAML: %s\n", err)
	}

	// open modem
	c := &serial.Config{
		Name: "/dev/tty.usbserial-A906XKUI",
		Baud: 19200,
	}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatalf("Could not open serial port: %s\n", err)
	}
	defer s.Close()
	modem = plm.New(plm.NewPort(s, 5 * time.Second), 5 * time.Second)
	defer modem.Close()

	for {
		// get the latest outdoor scene
		now := time.Now()
		outdoorScene, err := home.GetOutdoorScene(now)
		if err != nil {
			log.Printf("Unable to get OutdoorScene for this interval: %s\n", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		for _, light := range lighting.Lights {
			if light.Type != Dimmer.String() {
				log.Printf("Light '%s' is not a dimmer, skipping\n", light.Name)
				continue
			}

			desiredBrightnessPercent, err := light.DesiredBrightnessPercent(outdoorScene)
			if err != nil {
				log.Printf("Could not get desired brightness for light '%s': %s\n", light.Name, err)
				continue
			}

			log.Printf("Setting desired brightness for light '%s' to %f%%\n", light.Name, desiredBrightnessPercent)
			for _, address := range light.InsteonOptions.Addresses {
				addr := insteon.Address{}
				err := addr.UnmarshalText([]byte(address))
				if err != nil {
					log.Printf("Could not parse insteon address for light '%s', address %s: %s\n", light.Name, address, err)
					continue
				}

				device, err := modem.Network.Connect(addr)
				if err != nil {
					log.Printf("Could not connect to device at address %s: %s\n", address, err)
					continue
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
						continue
					}
					if currentOnLevel > 0 && ((!outdoorScene.AfterMeridian && currentOnLevel < onLevel) || (outdoorScene.AfterMeridian && currentOnLevel > onLevel)) {
						log.Printf("Adjusting current OnLevel of device at address %s to %d\n", address, onLevel)
						err = dimmer.OnAtRamp(onLevel, 60)
						if err != nil {
							log.Printf("Could not adjust OnLevel of device at address %s: %s\n", address, err)
							continue
						}
					}
				} else {
					log.Printf("Device at address %s is not a dimmer", address)
					continue
				}
			}
		}

		time.Sleep(15 * time.Minute)
	}
}

