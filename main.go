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
		log.Println("DEBUG")
		image, err := os.ReadFile(utils.CameraDebugFilename)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("READ")
		digits, err := camera.RecognizeDigits(image)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(digits)
		return
	}
	server.Start()
}
