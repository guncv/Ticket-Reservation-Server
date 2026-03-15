package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"
	"github.com/guncv/ticket-reservation-server/internal/service/event"
	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
)

type EventHandler struct {
	eventService event.EventService
	log          log.Logger
}

func NewEventHandler(
	eventService event.EventService,
	log log.Logger,
) *EventHandler {
	return &EventHandler{
		eventService: eventService,
		log:          log,
	}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req dto.CreateEventReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error(c.Request.Context(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.eventService.CreateEvent(c.Request.Context(), req)
	if err != nil {
		h.log.Error(c.Request.Context(), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, nil)
}

func (h *EventHandler) UpdateEvent(c *gin.Context) {
	var req dto.UpdateEventReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error(c.Request.Context(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.eventService.UpdateEvent(c.Request.Context(), req)
	if err != nil {
		h.log.Error(c.Request.Context(), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (h *EventHandler) GetAllEvents(c *gin.Context) {
	events, err := h.eventService.GetAllEvents(c.Request.Context())
	if err != nil {
		h.log.Error(c.Request.Context(), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

func (h *EventHandler) GetEventByID(c *gin.Context) {
	id := c.Param("id")

	idUUID, err := uuid.Parse(id)
	if err != nil {
		h.log.Error(c.Request.Context(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := h.eventService.GetEventByID(c.Request.Context(), idUUID)
	if err != nil {
		h.log.Error(c.Request.Context(), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

func (h *EventHandler) ReserveEventTicket(c *gin.Context) {
	var req dto.ReserveEventTicketReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error(c.Request.Context(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.eventService.ReserveEventTicket(c.Request.Context(), req)
	if err != nil {
		h.log.Error(c.Request.Context(), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
