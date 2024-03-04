package camera

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kabukky/homeautomation/pushover"
	"github.com/kabukky/homeautomation/utils"
)

var (
	globalTimesMutex      sync.RWMutex
	globalTimes           = createTimes()
	washingMachineRunning bool
	dryerRunning          bool
	maxInterpolatedCount  = 3
	updatePauseInMinutes  = 2
)

type Times struct {
	DryerMinutes                    int `json:"dryer_minutes"` // done: -1, off: -2
	DryerInterpolatedCount          int `json:"-"`
	WashingMachineMinutes           int `json:"washing_machine_minutes"` // done: -1, off: -2
	WashingMachineInterpolatedCount int `json:"-"`
}

func init() {
	if utils.CameraDebug {
		return
	}
	go func() {
		for {
			refreshTimes()
			globalTimesMutex.RLock()
			log.Println(fmt.Sprintf("globalTimes: %+v", globalTimes))
			globalTimesMutex.RUnlock()
			// Refresh every x minutes
			time.Sleep(time.Duration(updatePauseInMinutes) * time.Minute)
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
		image, err := GetImage(utils.CameraHostDryer)
		if err != nil {
			log.Println("Error while getting dryer image:", err)
			return
		}
		interpolated := false
		digits, err := RecognizeDryer(image)
		if err != nil {
			if err == errorNoDigits {
				// Off
			} else if err == errorMachineDone {
				times.DryerMinutes = -1
			} else {
				log.Println("Error while recognizing dryer digits:", err)
				// Interpolate if possible
				globalTimesMutex.RLock()
				minutesBefore := globalTimes.DryerMinutes
				interpolatedCount := globalTimes.DryerInterpolatedCount
				globalTimesMutex.RUnlock()
				if interpolatedCount < maxInterpolatedCount && minutesBefore > updatePauseInMinutes {
					times.DryerMinutes = minutesBefore - updatePauseInMinutes
					times.DryerInterpolatedCount = interpolatedCount + 1
					interpolated = true
				}
			}
		} else if len(digits) == 3 {
			// Compute minutes from digits
			times.DryerMinutes = 0
			times.DryerMinutes += 60 * digits[0]
			times.DryerMinutes += 10 * digits[1]
			times.DryerMinutes += digits[2]
		}
		if !interpolated {
			times.WashingMachineInterpolatedCount = 0
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
		interpolated := false
		digits, err := RecognizeWashingMachine(image)
		if err != nil {
			if err == errorNoDigits {
				// Off
			} else if err == errorMachineDone {
				times.WashingMachineMinutes = -1
			} else {
				log.Println("Error while recognizing washing machine digits:", err)
				// Interpolate if possible
				globalTimesMutex.RLock()
				minutesBefore := globalTimes.WashingMachineMinutes
				interpolatedCount := globalTimes.WashingMachineInterpolatedCount
				globalTimesMutex.RUnlock()
				if interpolatedCount < maxInterpolatedCount && minutesBefore > updatePauseInMinutes {
					times.WashingMachineMinutes = minutesBefore - updatePauseInMinutes
					times.WashingMachineInterpolatedCount = interpolatedCount + 1
					interpolated = true
				}
			}
		} else if len(digits) == 3 {
			// Compute minutes from digits
			times.WashingMachineMinutes = 0
			times.WashingMachineMinutes += 60 * digits[0]
			times.WashingMachineMinutes += 10 * digits[1]
			times.WashingMachineMinutes += digits[2]
		}
		if !interpolated {
			times.WashingMachineInterpolatedCount = 0
		}
	}()

	wg.Wait()

	// Check if one of the machines just got done
	globalTimesMutex.RLock()
	if times.DryerMinutes == -1 && globalTimes.DryerMinutes != -1 && dryerRunning {
		// Dryer was running last time and is now done
		log.Println("Dryer just finished!")
		go pushover.SendPush("Der Trockner ist fertig! 🥳")
	}
	if times.WashingMachineMinutes == -1 && globalTimes.WashingMachineMinutes != -1 && washingMachineRunning {
		// Washing machine was running last time and is now done
		log.Println("Washing machine just finished!")
		go pushover.SendPush("Die Waschmaschine ist fertig! 🤩")
	}
	globalTimesMutex.RUnlock()

	globalTimesMutex.Lock()
	if times.DryerMinutes > -1 {
		dryerRunning = true
	} else if times.DryerMinutes == -1 {
		dryerRunning = false
	}
	if times.WashingMachineMinutes > -1 {
		washingMachineRunning = true
	} else if times.WashingMachineMinutes == -1 {
		washingMachineRunning = false
	}
	globalTimes = times
	globalTimesMutex.Unlock()
}
