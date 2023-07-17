package grpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"telematics-generator/pkg/cache"
	"telematics-generator/pkg/models"
	"telematics-generator/protobuf"
	"testing"
	"time"
)

func TestGetLatestData(t *testing.T) {
	c := cache.NewTelematicsDataCache(10)
	s := NewServer(c)
	data := models.TelematicsData{
		VehicleID: 1,
		Timestamp: time.Now(),
		Speed:     10,
		Latitude:  50.4500,
		Longitude: 30.5233,
	}

	c.Add(data)

	req := &emptypb.Empty{}
	resp, err := s.GetLatestData(context.Background(), req)
	if err != nil {
		t.Fatalf("GetLatestData() error = %v", err)
	}

	if resp.VehicleId != int32(data.VehicleID) ||
		resp.Speed != int32(data.Speed) ||
		resp.Latitude != data.Latitude ||
		resp.Longitude != data.Longitude {
		t.Errorf("GetLatestData() got unexpected response")
	}
}

func TestGetRangeData(t *testing.T) {
	c := cache.NewTelematicsDataCache(10)
	s := NewServer(c)
	now := time.Now()
	data1 := models.TelematicsData{
		VehicleID: 1,
		Timestamp: now.Add(-5 * time.Second),
		Speed:     10,
		Latitude:  50.4500,
		Longitude: 30.5233,
	}

	data2 := models.TelematicsData{
		VehicleID: 1,
		Timestamp: time.Now(),
		Speed:     20,
		Latitude:  50.4510,
		Longitude: 30.5240,
	}

	c.Add(data1)
	c.Add(data2)

	req := &protobuf.RangeDataRequest{
		FromTimestamp: now.Add(-20 * time.Second).UnixNano(),
		ToTimestamp:   now.Add(20 * time.Second).UnixNano(),
	}
	stream := newMockTelematicsDataService_GetRangeDataServer()
	err := s.GetRangeData(req, stream)
	if err != nil {
		t.Fatalf("GetRangeData() error = %v", err)
	}

	if len(stream.responses) != 2 {
		t.Fatalf("GetRangeData() expected 2 responses, got %v", len(stream.responses))
	}
}

type mockTelematicsDataService_GetRangeDataServer struct {
	grpc.ServerStream
	responses []*protobuf.TelematicsDataProto
}

func newMockTelematicsDataService_GetRangeDataServer() *mockTelematicsDataService_GetRangeDataServer {
	return &mockTelematicsDataService_GetRangeDataServer{}
}

func (m *mockTelematicsDataService_GetRangeDataServer) Send(resp *protobuf.TelematicsDataProto) error {
	m.responses = append(m.responses, resp)
	return nil
}
