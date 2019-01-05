package main

import (
	"flag"
	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
	"github.com/tarm/serial"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"
)

var modem *plm.PLM

func main() {
	insteon.Log.Level(insteon.LevelTrace)
	log.Println("Hello")

	port := flag.String("port", "", "the path to the PLM")
	flag.Parse()
	if *port == "" {
		log.Fatalln("Port flag is required")
	}

	// parse home file
	home := &Location{}
	homeFile, err := ioutil.ReadFile("config/home.yaml")
	if err != nil {
		log.Fatalf("Unable to open home configuration file: %s\n", err)
	}
	err = yaml.UnmarshalStrict(homeFile, &home)
	if err != nil {
		log.Fatalf("Unable to unmarshal home YAML: %s\n", err)
	}

	// parse lighting file
	lighting := &Lighting{}

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
		Name: *port,
		Baud: 19200,
	}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatalf("Could not open serial port: %s\n", err)
	}
	defer s.Close()
	modem = plm.New(plm.NewPort(s, 5 * time.Second), 5 * time.Second)
	defer modem.Close()

	// Start lighting control loop
	lightControlDone := make(chan bool)
	lightControl := NewLightControl(home, lighting, modem)
	go lightControl.Start(5 * time.Minute, lightControlDone)


	// Wait for interrupt or thread death
	intChan := make(chan os.Signal, 1)
	signal.Notify(intChan, os.Interrupt)
	select {
	case <-lightControlDone:
		os.Exit(1)
	case <- intChan:
		log.Println("Interrupt caught!")
		os.Exit(0)
	}
}

