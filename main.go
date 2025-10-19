package main

import (
	"log"
	"runtime"

	"github.com/kabukky/homeautomation/camera"
	"github.com/kabukky/homeautomation/server"
	"github.com/kabukky/homeautomation/utils"
)

func main() {
	if utils.CameraDebug {
		runtime.LockOSThread()
		if utils.CameraDebugViaNetwork {
			// Dryer
			// image, err := camera.GetImageRtsp(utils.CameraHostDryer, utils.CameraUsername, utils.CameraPassword)
			// if err != nil {
			// 	log.Fatalln(err)
			// }
			// digits, err := camera.RecognizeDryer(*image)
			// if err != nil {
			// 	log.Fatalln(err)
			// }
			// Washing machine
			imageBytes, err := camera.GetImage(utils.CameraHostWashingMachine)
			if err != nil {
				log.Fatalln(err)
			}
			digits, err := camera.RecognizeWashingMachine(imageBytes)
			if err != nil {
				log.Fatalln(err)
			}

			log.Println(digits)
		}
		// digits, err := camera.RecognizeDryer(*image)
		// if err != nil {
		// 	log.Fatalln(err)
		// }
		// log.Println(digits)
		return
	}
	server.Start()
}
