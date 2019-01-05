package main

import (
	"flag"
	"github.com/abates/insteon/plm"
	log "github.com/sirupsen/logrus"
	"github.com/tarm/serial"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/signal"
	"time"
)

var modem *plm.PLM

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	log.Info("Initializing Halo")

	port := flag.String("port", "", "the path to the PLM")
	flag.Parse()
	if *port == "" {
		log.Fatal("Port flag is required")
	}

	// parse home file
	home := &Location{}
	homeFile, err := ioutil.ReadFile("config/home.yaml")
	if err != nil {
		log.Fatal("Unable to open home configuration file: ", err)
	}
	err = yaml.UnmarshalStrict(homeFile, &home)
	if err != nil {
		log.Fatal("Unable to unmarshal home YAML: ", err)
	}

	// parse lighting file
	lighting := &Lighting{}

	lightingFile, err := ioutil.ReadFile("config/lighting.yaml")
	if err != nil {
		log.Fatal("Unable to open lighting configuration file: ", err)
	}

	err = yaml.UnmarshalStrict(lightingFile, &lighting)
	if err != nil {
		log.Fatal("Unable to unmarshal lighting YAML: ", err)
	}

	// open modem
	c := &serial.Config{
		Name: *port,
		Baud: 19200,
	}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal("Could not open serial port: ", err)
	}
	defer s.Close()
	modem = plm.New(plm.NewPort(s, 5*time.Second), 5*time.Second)
	defer modem.Close()

	// Start lighting control loop
	lightControlDone := make(chan bool)
	lightControl := NewLightControl(home, lighting, modem)
	go lightControl.Start(5*time.Minute, lightControlDone)

	// Wait for interrupt or thread death
	intChan := make(chan os.Signal, 1)
	signal.Notify(intChan, os.Interrupt)
	select {
	case <-lightControlDone:
		os.Exit(1)
	case <-intChan:
		log.Debug("Interrupt caught!")
		os.Exit(0)
	}
}
