package repository

import (
	"time"

	"github.com/valorvig/bookings/internal/models"
)

// make some methods or functions available to the repository for the database - NewRepo returns *Repository
type DatabaseRepo interface {
	AllUsers() bool

	// get info from the reservation page and put them to the database
	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(r models.RoomRestriction) error
	SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error)
	SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error)
	GetRoomByID(id int) (models.Room, error)
}
