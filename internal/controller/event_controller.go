package controller

import (
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/service"
)

type EventController struct {
	eventService *service.EventService
}

func NewEventController(eventService *service.EventService) *EventController {
	return &EventController{
		eventService: eventService,
	}
}

func (c *EventController) CreateEvent(userID uint, req models.EventRequest) (*models.Event, error) {
	return c.eventService.CreateEvent(userID, req)
}

func (c *EventController) GetEvent(eventID uint) (*models.Event, error) {
	return c.eventService.GetEvent(eventID)
}

func (c *EventController) GetUserEvents(userID uint) ([]models.Event, error) {
	return c.eventService.GetUserEvents(userID)
}

func (c *EventController) UpdateEvent(eventID uint, userID uint, req models.UpdateEventRequest) (*models.Event, error) {
	return c.eventService.UpdateEvent(eventID, userID, req)
}

func (c *EventController) DeleteEvent(eventID uint, userID uint) error {
	return c.eventService.DeleteEvent(eventID, userID)
}

func (c *EventController) GetEventByURL(url string) (*models.Event, error) {
	return c.eventService.GetEventByURL(url)
}

func (c *EventController) CheckEventPassword(eventID uint, password string) error {
	return c.eventService.CheckEventPassword(eventID, password)
}
