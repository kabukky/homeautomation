package main

import (
	"log"
	"os"
	"runtime"

	"github.com/kabukky/homeautomation/camera"
	"github.com/kabukky/homeautomation/server"
	"github.com/kabukky/homeautomation/utils"
)

func main() {
	if utils.CameraDebug {
		runtime.LockOSThread()
		image, err := os.ReadFile(utils.CameraDebugFilename)
		if err != nil {
			log.Fatalln(err)
		}
		digits, err := camera.RecognizeWashingMachine(image)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(digits)
		return
	}
	server.Start()
}
