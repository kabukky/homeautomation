package main

import (
	"log"
	"runtime"

	"github.com/kabukky/homeautomation/camera"
	"github.com/kabukky/homeautomation/server"
	"github.com/kabukky/homeautomation/utils"
	"gocv.io/x/gocv"
)

func main() {
	if utils.CameraDebug {
		runtime.LockOSThread()
		var image *gocv.Mat
		var err error
		if utils.CameraDebugViaNetwork {
			image, err = camera.GetImageRtsp(utils.CameraHostDryer, utils.CameraUsername, utils.CameraPassword)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			// Removed this with Rtsp support.
			// TODO: Read gocv.Mat from image file.
			log.Fatalln("From image not implemented yet")
		}
		digits, err := camera.RecognizeDryer(*image)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(digits)
		return
	}
	server.Start()
}
