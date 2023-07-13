package generator

import (
	"testing"
	"time"
)

func TestNewRandomTelematicsGenerator(t *testing.T) {
	generator := NewRandomTelematicsGenerator(100, 10)
	if generator.maxSpeed != 100 {
		t.Errorf("expected MaxSpeed 100, but got %v", generator.maxSpeed)
	}
	if generator.maxTimeStep != 10 {
		t.Errorf("expected MaxTimeStep 10, but got %v", generator.maxTimeStep)
	}
}

func TestGenerate(t *testing.T) {
	gen := NewRandomTelematicsGenerator(100, 10)

	vehicleID := 99
	telematics := gen.Generate(vehicleID)

	select {
	case data := <-telematics:
		if data.VehicleID != 99 {
			t.Errorf("vehicleID not correct, expected %d, got %v", 99, data.VehicleID)
		}

		if data.Speed > gen.maxSpeed {
			t.Errorf("expected speed greater than maxSpeed %v, got %v", gen.maxSpeed, data.Speed)
		}

		if data.Speed < 0 {
			t.Errorf("expected speed less than 0, got %v", data.Speed)
		}

		if data.Latitude < -90 || data.Latitude > 90 {
			t.Errorf("latitude out of range, expected value between -90 and 90, got %v", data.Latitude)
		}

		if data.Longitude < -180 || data.Longitude > 180 {
			t.Errorf("longitude out of range, expected value between -180 and 180, got %v", data.Longitude)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for telematics data")
	}
}
