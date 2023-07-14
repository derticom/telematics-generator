package cache

import (
	"telematics-generator/pkg/models"
	"testing"
	"time"
)

func TestAddAndGetLatest(t *testing.T) {
	c := NewTelematicsDataCache(10)
	data := models.TelematicsData{
		VehicleID: 1,
		Timestamp: time.Now(),
		Speed:     10,
		Latitude:  50.4500,
		Longitude: 30.5233,
	}

	c.Add(data)

	latest, ok := c.GetLatest()
	if !ok || latest != data {
		t.Errorf("GetLatest() = %v, want %v", latest, data)
	}
}

func TestCapacity(t *testing.T) {
	c := NewTelematicsDataCache(1)
	data1 := models.TelematicsData{
		VehicleID: 1,
		Timestamp: time.Now(),
		Speed:     10,
		Latitude:  50.4500,
		Longitude: 30.5233,
	}
	data2 := models.TelematicsData{
		VehicleID: 1,
		Timestamp: time.Now(),
		Speed:     20,
		Latitude:  50.4501,
		Longitude: 30.5234,
	}

	c.Add(data1)
	c.Add(data2)

	latest, ok := c.GetLatest()
	if !ok || latest != data2 {
		t.Errorf("GetLatest() = %v, want %v", latest, data2)
	}
}

func TestGetRange(t *testing.T) {
	c := NewTelematicsDataCache(10)
	now := time.Now()
	data1 := models.TelematicsData{
		VehicleID: 1,
		Timestamp: now.Add(-5 * time.Minute),
		Speed:     10,
		Latitude:  50.4500,
		Longitude: 30.5233,
	}
	data2 := models.TelematicsData{
		VehicleID: 1,
		Timestamp: now,
		Speed:     20,
		Latitude:  50.4501,
		Longitude: 30.5234,
	}

	c.Add(data1)
	c.Add(data2)

	from := now.Add(-10 * time.Minute)
	to := now.Add(-1 * time.Minute)
	result, err := c.GetRange(from, to)
	if err != nil {
		t.Errorf("GetRange() error = %v", err)
		return
	}
	if len(result) != 1 || result[0] != data1 {
		t.Errorf("GetRange() = %v, want %v", result, []models.TelematicsData{data1})
	}
}
