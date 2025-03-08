package handlers

import (
	"agent-backend/src/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

var userLocations = []models.UserLocation{
	{UserID: "1", Latitude: 37.7749, Longitude: -122.4194, Timestamp: "2025-02-21T10:00:00Z"},
	{UserID: "2", Latitude: 34.0522, Longitude: -118.2437, Timestamp: "2025-02-21T11:00:00Z"},
}

func GetSharedLocations(c echo.Context) error {
	// userId := c.Param("userId")
	// Fetch shared locations for the given userId from the database
	// For example:
	// locations, err := fetchSharedLocationsFromDB(userId)
	// if err != nil {
	//     return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	// }
	// return c.JSON(http.StatusOK, locations)

	// Placeholder response
	return c.JSON(http.StatusOK, userLocations)
}

func UpdateUserLocation(c echo.Context) error {
	var newUserLocation models.UserLocation
	if err := c.Bind(&newUserLocation); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid input"})
	}
	userLocations = append(userLocations, newUserLocation)
	return c.JSON(http.StatusCreated, newUserLocation)
}
