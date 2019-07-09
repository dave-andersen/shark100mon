package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/goburrow/modbus"
	"log"
	"math"
	"net/http"
	"sync"
	"time"
)

const (
	WATTS_ALPHA = 0.15
	POLLING_MS  = 1000
)

var interactive = flag.Bool("i", false, "output to console interactively")

var wattsEwma float32 = 0.0
var voltsGlobal float32 = 0.0
var frequencyGlobal float32 = 0.0
var wattsMutex = sync.Mutex{}

type updaterFn func(watts float32, volts float32, frequency float32)

func updatePower(watts float32, volts float32, frequency float32) {
	wattsMutex.Lock()
	defer wattsMutex.Unlock()
	if wattsEwma == 0.0 {
		wattsEwma = watts
	}
	wattsEwma = wattsEwma*(1.0-WATTS_ALPHA) + watts*WATTS_ALPHA
	voltsGlobal = volts
	frequencyGlobal = frequency
}

func printPower(watts float32, volts float32, frequency float32) {
	updatePower(watts, volts, frequency)
	wattsMutex.Lock()
	defer wattsMutex.Unlock()
	fmt.Printf("Watts: %d  volts: %.2f   frequency: %.4f hz\n",
		int(wattsEwma), volts, frequency)
}

func readFloatRegisters(client modbus.Client, regStart uint16, length uint16) (float32, error) {
	results, err := client.ReadHoldingRegisters(regStart, length)
	var res float32 = 0.0
	if err == nil {
		resint := binary.BigEndian.Uint32(results)
		res = math.Float32frombits(resint)
	}
	return res, err
}

func readPowerLoopInternal(client modbus.Client, f updaterFn) {
	for {
		watts, err := readFloatRegisters(client, 0x383, 2)
		if err != nil {
			f(0, 0, 0)
			log.Println("Error:", err)
			return
		}
		volts, err := readFloatRegisters(client, 0x03ED, 2)
		if err != nil {
			f(0, 0, 0)
			log.Println("Error:", err)
			return
		}

		frequency, err := readFloatRegisters(client, 0x0401, 2)
		if err != nil {
			f(0, 0, 0)
			log.Println("Error:", err)
			return
		}

		f(watts, volts, frequency)
		time.Sleep(POLLING_MS * time.Millisecond)
	}
}

func readPowerLoop(f updaterFn) {
	wattsEwma = 0
	// The modbus connection times out periodically, so
	// reestablish when needed
	for {
		client := modbus.TCPClient("192.168.49.5:502")
		readPowerLoopInternal(client, f)
		time.Sleep(1 * time.Second)
	}
}

func getPower() (float32, float32, float32) {
	wattsMutex.Lock()
	defer wattsMutex.Unlock()
	return wattsEwma, voltsGlobal, frequencyGlobal
}

func main() {
	flag.Parse()

	if *interactive {
		readPowerLoop(printPower)
	} else {
		go readPowerLoop(updatePower)
		http.HandleFunc("/power", func(w http.ResponseWriter, r *http.Request) {
			watts, volts, frequency := getPower()
			fmt.Fprintf(w, "{\"watts\": %.0f, \"volts\": %.2f, \"frequency\": %.4f}\n", watts, volts, frequency)
		})
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "powermon.html")
		})
		log.Fatal(http.ListenAndServe(":8081", nil))
	}

}
