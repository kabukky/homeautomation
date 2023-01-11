package camera

import (
	"log"
	"sync"

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
	var wg sync.WaitGroup

	// Dryer
	wg.Add(1)
	go func() {
		defer wg.Done()
		image, err := GetImage(utils.CameraHostDryer)
		if err != nil {
			log.Println("Error while getting dryer image:", err)
			return
		}
		digits, err := RecognizeDryer(image)
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
	}()

	// Washing Machine
	wg.Add(1)
	go func() {
		defer wg.Done()
		image, err := GetImage(utils.CameraHostWashingMachine)
		if err != nil {
			log.Println("Error while getting washing machine image:", err)
			return
		}
		digits, err := RecognizeWashingMachine(image)
		if err != nil {
			if err == errorMachineDone {
				times.WashingMachineMinutes = -1
			} else {
				log.Println("Error while recognizing washing machine digits:", err)
			}
		} else if len(digits) == 3 {
			times.WashingMachineMinutes = 0
			times.WashingMachineMinutes += 60 * digits[0]
			times.WashingMachineMinutes += 10 * digits[1]
			times.WashingMachineMinutes += digits[2]
		}
	}()

	wg.Wait()
	return &times, nil
}
