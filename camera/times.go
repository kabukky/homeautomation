package camera

import (
	"log"
	"sync"
	"time"

	"github.com/kabukky/homeautomation/utils"
)

var (
	globalTimesMutex sync.RWMutex
	globalTimes      = createTimes()
)

type Times struct {
	DryerMinutes          int `json:"dryer_minutes"`           // done: -1, off: -2
	WashingMachineMinutes int `json:"washing_machine_minutes"` // done: -1, off: -2
}

func init() {
	if utils.CameraDebug {
		return
	}
	go func() {
		for {
			refreshTimes()
			globalTimesMutex.RLock()
			log.Println("globalTimes:", globalTimes)
			globalTimesMutex.RUnlock()
			// Refresh every minute
			time.Sleep(1 * time.Minute)
		}
	}()
}

func createTimes() Times {
	times := Times{
		DryerMinutes:          -2,
		WashingMachineMinutes: -2,
	}
	return times
}

func GetTimes() *Times {
	globalTimesMutex.RLock()
	times := globalTimes
	globalTimesMutex.RUnlock()
	return &times
}

func refreshTimes() {
	times := createTimes()
	var wg sync.WaitGroup

	// Dryer
	wg.Add(1)
	go func() {
		defer wg.Done()
		image, err := getImage(utils.CameraHostDryer)
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
		image, err := getImage(utils.CameraHostWashingMachine)
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

	// Check if one of the machines just got done
	globalTimesMutex.RLock()
	lastTimes := globalTimes
	globalTimesMutex.RUnlock()
	if times.DryerMinutes == -1 && lastTimes.DryerMinutes != -2 && lastTimes.DryerMinutes != -1 {
		// Dryer was running last time and is now done
		log.Println("Dryer just finished!")
	}
	if times.WashingMachineMinutes == -1 && lastTimes.WashingMachineMinutes != -2 && lastTimes.WashingMachineMinutes != -1 {
		// Washing machine was running last time and is now done
		log.Println("Washing machine just finished!")
	}

	globalTimesMutex.Lock()
	globalTimes = times
	globalTimesMutex.Unlock()
}
