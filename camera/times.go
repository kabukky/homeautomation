package camera

import (
	"log"

	"github.com/kabukky/homeautomation/utils"
)

type Times struct {
	DryerMinutes          int `json:"dryer_minutes"`           // done: -1, off: -2
	WashingMachineMinutes int `json:"washing_machine_minutes"` // done: -1, off: -2
}

func GetTimes() (*Times, error) {
	times := Times{
		DryerMinutes:          -2,
		WashingMachineMinutes: -2,
	}
	// Dryer
	image, err := GetImage(utils.CameraHostDryer)
	if err != nil {
		return nil, err
	}
	digits, err := RecognizeDigits(image)
	if err != nil {
		if err == errorMachineDone {
			times.DryerMinutes = -1
		} else {
			log.Println("Error while recognizing dryer digits:", err)
		}
	} else if len(digits) == 3 {
		times.DryerMinutes = 0
		times.DryerMinutes += 60 * digits[0]
		times.DryerMinutes += 10 * digits[1]
		times.DryerMinutes += digits[2]
	}

	return &times, nil
}
