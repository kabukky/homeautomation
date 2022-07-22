package calendar

import (
	"context"
	"log"
	"time"

	"github.com/kabukky/homeautomation/utils"
	"google.golang.org/api/calendar/v3"
)

type Event struct {
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

var (
	service *calendar.Service
)

func init() {
	s, err := calendar.NewService(context.Background())
	if err != nil {
		log.Fatal("Could not create google calendar service:", err)
	}
	for _, calendarID := range utils.GoogleCalendarIDs {
		if calendarID == "" {
			continue
		}
		_, err = s.CalendarList.Insert(&calendar.CalendarListEntry{Id: calendarID}).Do()
		if err != nil {
			log.Fatal("Could not add google calendar", calendarID, ":", err)
		}
	}
	service = s
}

func GetEvents() (map[string][]Event, error) {
	// TODO: Implement rate limiting.
	// We probably only need to poll every few minutes and use cached events in-between.
	calendars := make(map[string][]Event)
	for _, calendarID := range utils.GoogleCalendarIDs {
		events := make([]Event, 0)
		t := time.Now().Format(time.RFC3339)
		googleEvents, err := service.Events.List(calendarID).ShowDeleted(false).
			SingleEvents(true).TimeMin(t).MaxResults(30).OrderBy("startTime").Do()
		if err != nil {
			return nil, err
		}
		for _, item := range googleEvents.Items {
			var parsedStartDate time.Time
			var parsedEndDate time.Time
			if item.Start.DateTime != "" {
				parsedStartDate, err = time.Parse(time.RFC3339, item.Start.DateTime)
				if err != nil {
					return nil, err
				}
			} else {
				parsedStartDate, err = time.ParseInLocation("2006-01-02", item.Start.Date, time.Local)
				if err != nil {
					return nil, err
				}
			}
			if item.End.DateTime != "" {
				parsedEndDate, err = time.Parse(time.RFC3339, item.End.DateTime)
				if err != nil {
					return nil, err
				}
			} else {
				parsedEndDate, err = time.ParseInLocation("2006-01-02", item.End.Date, time.Local)
				if err != nil {
					return nil, err
				}
			}
			events = append(events, Event{Name: item.Summary, StartDate: parsedStartDate, EndDate: parsedEndDate})
		}
		calendars[calendarID] = events
	}
	return calendars, nil
}
