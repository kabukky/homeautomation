package camera

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"log"
	"sort"
	"strconv"

	"github.com/kabukky/homeautomation/utils"
	"gocv.io/x/gocv"
)

type digit struct {
	xPos int
	mat  *gocv.Mat
}

type digitKey struct {
	top         bool
	topLeft     bool
	topRight    bool
	center      bool
	bottomLeft  bool
	bottomRight bool
	bottom      bool
}

var (
	digitDefinitions = map[digitKey]int{
		{true, true, true, false, true, true, true}:      0,
		{false, false, true, false, false, true, false}:  1, // Maybe other way around?
		{true, false, true, true, true, false, true}:     2,
		{true, false, true, true, false, true, true}:     3,
		{false, true, true, true, false, true, false}:    4,
		{true, true, false, true, false, true, true}:     5,
		{true, true, false, true, true, true, true}:      6,
		{true, false, true, false, false, true, false}:   7, // Maybe different?
		{true, true, true, true, true, true, true}:       8,
		{true, true, true, true, false, true, true}:      9,
		{false, false, false, true, false, false, false}: -1, // Means dash (-)
	}
	whiteThreshold   = 0.7
	errorMachineDone = errors.New("machine done")
	errorNoDigits    = errors.New("no digits found")
)

func RecognizeDryer(mat gocv.Mat) ([]int, error) {
	defer mat.Close()

	// Rotate
	rotated := gocv.NewMat()
	defer rotated.Close()
	gocv.Rotate(mat, &rotated, gocv.Rotate180Clockwise)

	// Coordinates for perspective transform
	displayCoords := []image.Point{
		image.Point{1540, 1047}, // top-left
		image.Point{1540, 1097}, // bottom-left
		image.Point{1655, 1098}, // bottom-right
		image.Point{1658, 1045}, // top-right
	}
	return recognizeDigits(&rotated, displayCoords, 210)
}

func RecognizeWashingMachine(imageBytes []byte) ([]int, error) {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, err
	}
	mat, err := gocv.ImageToMatRGB(img)
	if err != nil {
		return nil, err
	}
	defer mat.Close()

	// Rotate
	rotated := gocv.NewMat()
	defer rotated.Close()
	gocv.Rotate(mat, &rotated, gocv.Rotate90CounterClockwise)

	// Coordinates for perspective transform
	displayCoords := []image.Point{
		image.Point{437, 215}, // top-left
		image.Point{442, 258}, // bottom-left
		image.Point{517, 246}, // bottom-right
		image.Point{512, 201}, // top-right
	}
	return recognizeDigits(&rotated, displayCoords, 160)
}

func recognizeDigits(mat *gocv.Mat, displayCoords []image.Point, minThreshold float32) ([]int, error) {
	if utils.CameraDebug {
		temp := mat.Clone()
		defer temp.Close()
		gocv.Polylines(&temp, gocv.NewPointsVectorFromPoints([][]image.Point{displayCoords}), true, color.RGBA{255, 0, 0, 255}, 1)
		window := gocv.NewWindow("Hello")
		for {
			window.IMShow(temp)
			if window.WaitKey(1) >= 0 {
				break
			}
		}
	}
	// Transform for correct perspective
	height := 220
	width := 400
	newImg := []image.Point{
		image.Point{0, 0},
		image.Point{0, height},
		image.Point{width, height},
		image.Point{width, 0},
	}
	transform := gocv.GetPerspectiveTransform(gocv.NewPointVectorFromPoints(displayCoords), gocv.NewPointVectorFromPoints(newImg))
	defer transform.Close()
	perspective := gocv.NewMat()
	defer perspective.Close()
	gocv.WarpPerspective(*mat, &perspective, transform, image.Point{width, height})
	if utils.CameraDebug {
		window := gocv.NewWindow("Hello")
		for {
			window.IMShow(perspective)
			if window.WaitKey(1) >= 0 {
				break
			}
		}
	}

	// Greyscale
	destGray := gocv.NewMat()
	defer destGray.Close()
	gocv.CvtColor(perspective, &destGray, gocv.ColorBGRToGray)

	// Threshold to make it black/white
	destThreshold := gocv.NewMat()
	defer destThreshold.Close()
	gocv.Threshold(destGray, &destThreshold, minThreshold, 255, gocv.ThresholdBinary)

	if utils.CameraDebug {
		window := gocv.NewWindow("Hello")
		for {
			window.IMShow(destThreshold)
			if window.WaitKey(1) >= 0 {
				break
			}
		}
	}

	// Get edges
	destEdge := gocv.NewMat()
	defer destEdge.Close()
	gocv.Canny(destThreshold, &destEdge, 50, 200)
	if utils.CameraDebug {
		window := gocv.NewWindow("Hello")
		for {
			window.IMShow(destEdge)
			if window.WaitKey(1) >= 0 {
				break
			}
		}
	}

	// Dilate edges
	destDilate := gocv.NewMat()
	defer destDilate.Close()
	kernelDilate := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernelDilate.Close()
	gocv.Dilate(destEdge, &destDilate, kernelDilate)
	if utils.CameraDebug {
		window := gocv.NewWindow("Hello")
		for {
			window.IMShow(destDilate)
			if window.WaitKey(1) >= 0 {
				break
			}
		}
	}

	destErode := gocv.NewMat()
	defer destErode.Close()
	kernelErode := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernelErode.Close()
	gocv.Erode(destDilate, &destErode, kernelErode)
	if utils.CameraDebug {
		window := gocv.NewWindow("Hello")
		for {
			window.IMShow(destErode)
			if window.WaitKey(1) >= 0 {
				break
			}
		}
	}

	// Get contours
	contours := gocv.FindContours(destErode, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()
	if contours.IsNil() {
		return nil, errors.New("could not find any contous")
	}

	// Sort
	points := contours.ToPoints()
	sortContoursBySize(points)
	var digits []digit
	for index, points := range points {
		if index > 2 {
			// TODO: Check if this is the second part of a disconnected "1"
			break
		}
		c := gocv.NewPointVectorFromPoints(points)
		defer c.Close()
		rect := gocv.BoundingRect(c)
		region := destThreshold.Region(rect)
		defer region.Close()
		digits = append(digits, digit{xPos: rect.Min.X, mat: &region})
	}
	// Sort left to right
	sort.Slice(digits, func(i, j int) bool {
		return digits[i].xPos < digits[j].xPos
	})
	if len(digits) == 0 {
		return nil, errorNoDigits
	}
	if len(digits) != 3 {
		return nil, errors.New("found wrong number of digits:" + strconv.Itoa(len(digits)))
	}

	// Recognize digits
	var recognizedDigits []int
	for index, digit := range digits {
		if utils.CameraDebug {
			window := gocv.NewWindow("Hello")
			for {
				window.IMShow(*digit.mat)
				if window.WaitKey(1) >= 0 {
					break
				}
			}
		}
		var skipRecognition bool
		var digitKey digitKey
		width := digit.mat.Cols()
		height := digit.mat.Rows()
		// Check if 1 or - (dash)
		// then most of the image is white as the outline takes most of the cutout
		allWhite := gocv.CountNonZero(*digit.mat)
		allTotal := width * height
		ratio := float64(width) / float64(height)
		if float64(allWhite)/float64(allTotal) > 0.6 {
			// Check ratio to see if this is a 1 or a -
			if ratio < 1 {
				// Ratio should be pretty thin for a 1
				if ratio < 0.37 {
					digitKey.topRight = true
					digitKey.bottomRight = true
					skipRecognition = true
				}
			} else {
				digitKey.center = true
				skipRecognition = true
			}

		}
		if !skipRecognition {
			// top
			topRect := image.Rect(0+(width/4), 0+(height/20), width-(width/4), height/5)
			topMat := digit.mat.Region(topRect)
			defer topMat.Close()
			topWhite := gocv.CountNonZero(topMat)
			topTotal := topRect.Dx() * topRect.Dy()
			if float64(topWhite)/float64(topTotal) > whiteThreshold {
				digitKey.top = true
			}
			if utils.CameraDebug {
				log.Println("top on:", float64(topWhite)/float64(topTotal))
				gocv.Rectangle(digit.mat, topRect, color.RGBA{0, 0, 0, 255}, 1)
				window := gocv.NewWindow("Hello")
				for {
					window.IMShow(*digit.mat)
					if window.WaitKey(1) >= 0 {
						break
					}
				}
			}
			// top left
			topLeftRect := image.Rect(0+(width/20), 0+(height/10), width/4, height/2)
			topLeftMat := digit.mat.Region(topLeftRect)
			defer topLeftMat.Close()
			topLeftWhite := gocv.CountNonZero(topLeftMat)
			topLeftTotal := topLeftRect.Dx() * topLeftRect.Dy()
			if float64(topLeftWhite)/float64(topLeftTotal) > whiteThreshold {
				digitKey.topLeft = true
			}
			if utils.CameraDebug {
				log.Println("top left on:", float64(topLeftWhite)/float64(topLeftTotal))
				gocv.Rectangle(digit.mat, topLeftRect, color.RGBA{0, 0, 0, 255}, 1)
				window := gocv.NewWindow("Hello")
				for {
					window.IMShow(*digit.mat)
					if window.WaitKey(1) >= 0 {
						break
					}
				}
			}
			// top right
			topRightRect := image.Rect(int(float64(width)*0.75), 0+(height/10), width-(width/20), height/2)
			topRightMat := digit.mat.Region(topRightRect)
			defer topRightMat.Close()
			topRightWhite := gocv.CountNonZero(topRightMat)
			topRightTotal := topRightRect.Dx() * topRightRect.Dy()
			if float64(topRightWhite)/float64(topRightTotal) > whiteThreshold {
				digitKey.topRight = true
			}
			if utils.CameraDebug {
				log.Println("top right on:", float64(topRightWhite)/float64(topRightTotal))
				gocv.Rectangle(digit.mat, topRightRect, color.RGBA{0, 0, 0, 255}, 1)
				window := gocv.NewWindow("Hello")
				for {
					window.IMShow(*digit.mat)
					if window.WaitKey(1) >= 0 {
						break
					}
				}
			}
			// center
			centerRect := image.Rect(0+(width/4), (height/2)-((height/2)/6), width-(width/4), (height/2)+((height/2)/6))
			centerMat := digit.mat.Region(centerRect)
			defer centerMat.Close()
			centerWhite := gocv.CountNonZero(centerMat)
			centerTotal := centerRect.Dx() * centerRect.Dy()
			if float64(centerWhite)/float64(centerTotal) > whiteThreshold {
				digitKey.center = true
			}
			if utils.CameraDebug {
				log.Println("center on:", float64(centerWhite)/float64(centerTotal))
				gocv.Rectangle(digit.mat, centerRect, color.RGBA{0, 0, 0, 255}, 1)
				window := gocv.NewWindow("Hello")
				for {
					window.IMShow(*digit.mat)
					if window.WaitKey(1) >= 0 {
						break
					}
				}
			}
			// bottom left
			bottomLeftRect := image.Rect(0+(width/20), height/2, width/4, height-(height/10))
			bottomLeftMat := digit.mat.Region(bottomLeftRect)
			defer bottomLeftMat.Close()
			bottomLeftWhite := gocv.CountNonZero(bottomLeftMat)
			bottomLeftTotal := bottomLeftRect.Dx() * bottomLeftRect.Dy()
			if float64(bottomLeftWhite)/float64(bottomLeftTotal) > whiteThreshold {
				digitKey.bottomLeft = true
			}
			if utils.CameraDebug {
				log.Println("bottom left on:", float64(bottomLeftWhite)/float64(bottomLeftTotal))
				gocv.Rectangle(digit.mat, bottomLeftRect, color.RGBA{0, 0, 0, 255}, 1)
				window := gocv.NewWindow("Hello")
				for {
					window.IMShow(*digit.mat)
					if window.WaitKey(1) >= 0 {
						break
					}
				}
			}
			// bottom right
			bottomRightRect := image.Rect(int(float64(width)*0.75), height/2, width-(width/20), height-(height/10))
			bottomRightMat := digit.mat.Region(bottomRightRect)
			defer bottomRightMat.Close()
			bottomRightWhite := gocv.CountNonZero(bottomRightMat)
			bottomRightTotal := bottomRightRect.Dx() * bottomRightRect.Dy()
			if float64(bottomRightWhite)/float64(bottomRightTotal) > whiteThreshold {
				digitKey.bottomRight = true
			}
			if utils.CameraDebug {
				log.Println("bottom right on:", float64(bottomRightWhite)/float64(bottomRightTotal))
				gocv.Rectangle(digit.mat, bottomRightRect, color.RGBA{0, 0, 0, 255}, 1)
				window := gocv.NewWindow("Hello")
				for {
					window.IMShow(*digit.mat)
					if window.WaitKey(1) >= 0 {
						break
					}
				}
			}
			// bottom
			bottomRect := image.Rect(0+(width/4), int(float64(height)*0.8), width-(width/4), height-(height/20))
			bottomMat := digit.mat.Region(bottomRect)
			defer bottomMat.Close()
			bottomWhite := gocv.CountNonZero(bottomMat)
			bottomTotal := bottomRect.Dx() * bottomRect.Dy()
			if float64(bottomWhite)/float64(bottomTotal) > whiteThreshold {
				digitKey.bottom = true
			}
			if utils.CameraDebug {
				log.Println("bottom on:", float64(bottomWhite)/float64(bottomTotal))
				gocv.Rectangle(digit.mat, bottomRect, color.RGBA{0, 0, 0, 255}, 1)
				window := gocv.NewWindow("Hello")
				for {
					window.IMShow(*digit.mat)
					if window.WaitKey(1) >= 0 {
						break
					}
				}
			}
		}
		recognizedDigit, ok := digitDefinitions[digitKey]
		if !ok {
			return nil, errors.New("could not recognize digit at index " + strconv.Itoa(index) + fmt.Sprintf(" for definition %+v", digitKey))
		}
		recognizedDigits = append(recognizedDigits, recognizedDigit)
	}

	if recognizedDigits[0] == -1 && recognizedDigits[1] == 0 && recognizedDigits[2] == -1 {
		return nil, errorMachineDone
	}

	return recognizedDigits, nil
}

func sortContoursBySize(contours [][]image.Point) {
	sort.Slice(contours, func(i, j int) bool {
		veci := gocv.NewPointVectorFromPoints(contours[i])
		defer veci.Close()
		recti := gocv.BoundingRect(veci)
		vecj := gocv.NewPointVectorFromPoints(contours[j])
		defer vecj.Close()
		rectj := gocv.BoundingRect(vecj)
		return (recti.Size().X * recti.Size().Y) > (rectj.Size().X * rectj.Size().Y)
	})
}
