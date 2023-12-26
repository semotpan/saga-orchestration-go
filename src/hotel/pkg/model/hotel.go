package model

import (
	"go.example/saga/pkg/jsonmap"
	"go.example/saga/pkg/store/postgres"
	"time"
)

// Hotel
type Hotel struct {
	ID           HotelID
	CreationTime time.Time
	Name         string
	Address      string
	Location     string
}

// HotelID
type HotelID int32

// Room
type Room struct {
	ID           RoomID
	CreationTime time.Time
	Name         string
	Floor        int8
	Number       int8
	Available    bool
	HotelID      HotelID
}

// RoomID
type RoomID int32

// RoomBookingEvent
type RoomBookingEvent struct {
	EventID   string       `json:"eventId"`
	MsgID     string       `json:"msgId"`
	Timestamp time.Time    `json:"timestamp"`
	Payload   EventPayload `json:"payload"`
}

// EventPayload
type EventPayload struct {
	HotelID   HotelID            `json:"hotelId"`
	RoomID    RoomID             `json:"roomId"`
	StartDate string             `json:"startDate"`
	EndDate   string             `json:"endDate"`
	Name      string             `json:"name"`
	Type      postgres.EventType `json:"type"`
}

// BookingStatus
type BookingStatus string

const (
	BookingStatusBooked    = "BOOKED"
	BookingStatusRejected  = "REJECTED"
	BookingStatusCancelled = "CANCELLED"
)

func (status BookingStatus) ToJSONMap() jsonmap.JSONMap {
	return jsonmap.JSONMap{"status": string(status)}
}
