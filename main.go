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
		var image []byte
		var err error
		if utils.CameraDebugViaNetwork {
			image, err = camera.GetImage(utils.CameraHostDryer)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			image, err = os.ReadFile(utils.CameraDebugFilename)
			if err != nil {
				log.Fatalln(err)
			}
		}
		digits, err := camera.RecognizeDryer(image)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(digits)
		return
	}
	server.Start()
}
