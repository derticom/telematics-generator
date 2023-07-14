package generator

import (
	geo "github.com/kellydunn/golang-geo"
	"math/rand"
	"telematics-generator/pkg/models"
	"time"
)

type Generator interface {
	Generate(vehicleID int) <-chan models.TelematicsData
}

type RandomTelematicsGenerator struct {
	maxSpeed    int
	maxTimeStep int
}

func NewRandomTelematicsGenerator(maxSpeedArg int, maxTimeStepArg int) *RandomTelematicsGenerator {
	return &RandomTelematicsGenerator{
		maxSpeed:    maxSpeedArg,
		maxTimeStep: maxTimeStepArg,
	}
}

func (g *RandomTelematicsGenerator) Generate(vehicleID int, stop chan struct{}) <-chan models.TelematicsData {
	out := make(chan models.TelematicsData)

	go func() {
		latitude := rand.Float64()*180 - 90
		longitude := rand.Float64()*360 - 180

		for {
			select {
			case <-stop:
				close(out)
				return
			default:
				deltaTime := rand.Float64() * float64(g.maxTimeStep)

				speed := rand.Intn(g.maxSpeed)

				distance := float64(speed) * (deltaTime / 3600)

				direction := rand.Float64() * 360

				p := geo.NewPoint(latitude, longitude)

				newPoint := p.PointAtDistanceAndBearing(distance, direction)

				out <- models.TelematicsData{
					VehicleID: vehicleID,
					Timestamp: time.Now(),
					Speed:     speed,
					Latitude:  newPoint.Lat(),
					Longitude: newPoint.Lng(),
				}

				time.Sleep(time.Duration(deltaTime) * time.Second)

				latitude = newPoint.Lat()
				longitude = newPoint.Lng()
			}
		}
	}()

	return out
}
