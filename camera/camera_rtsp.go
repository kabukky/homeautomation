package camera

import (
	"errors"
	"fmt"

	"gocv.io/x/gocv"
)

func GetImageRtsp(rtspCamHost string, username string, password string) (*gocv.Mat, error) {
	webcam, err := gocv.OpenVideoCapture(fmt.Sprintf("rtsp://%s:%s@%s:554/stream1", username, password, rtspCamHost))
	if err != nil {
		return nil, err
	}
	defer webcam.Close()

	img := gocv.NewMat()

	if ok := webcam.Read(&img); !ok {
		return nil, errors.New("could not read into mat")
	}
	if img.Empty() {
		return nil, errors.New("no image")
	}
	return &img, nil
}
