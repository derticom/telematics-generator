package test

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"
	"telematics-generator/pkg/cache"
	"telematics-generator/pkg/generator"
	"telematics-generator/protobuf"
)

type mockKafkaProducer struct {
	messages []*protobuf.TelematicsDataProto
}

func (m *mockKafkaProducer) ProduceMessage(data *protobuf.TelematicsDataProto) error {
	m.messages = append(m.messages, data)
	return nil
}

type mockServer struct {
	data []*protobuf.TelematicsDataProto
}

func (m *mockServer) GetLatestData(ctx context.Context, req *emptypb.Empty) (*protobuf.TelematicsDataProto, error) {
	if len(m.data) == 0 {
		return nil, status.Error(codes.NotFound, "no data available")
	}
	return m.data[len(m.data)-1], nil
}

func (m *mockServer) GetRangeData(ctx context.Context, req *protobuf.RangeDataRequest, srv protobuf.TelematicsDataService_GetRangeDataServer) error {
	if len(m.data) == 0 {
		return status.Error(codes.NotFound, "no data available")
	}

	for _, data := range m.data {
		if data.Timestamp >= req.FromTimestamp && data.Timestamp <= req.ToTimestamp {
			if err := srv.Send(data); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestGeneratorKafkaGRPC(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaProducer := &mockKafkaProducer{}

	gen := generator.NewRandomTelematicsGenerator(180, 5)

	dataCache := cache.NewTelematicsDataCache(100)

	stop := make(chan struct{})
	go func() {
		dataCh := gen.Generate(1, stop)
		for data := range dataCh {
			err := kafkaProducer.ProduceMessage(&protobuf.TelematicsDataProto{
				VehicleId: int32(data.VehicleID),
				Timestamp: data.Timestamp.UnixNano(),
				Speed:     int32(data.Speed),
				Latitude:  data.Latitude,
				Longitude: data.Longitude,
			})
			assert.NoError(t, err)
			fmt.Println(data)
			dataCache.Add(data)

		}
	}()

	time.Sleep(10 * time.Second)

	server := &mockServer{
		data: kafkaProducer.messages,
	}

	data, err := server.GetLatestData(ctx, &emptypb.Empty{})
	fmt.Printf("Data from GetLatestData: %v\n", data)
	assert.NoError(t, err)
	assert.NotNil(t, data)

	rangeReq := &protobuf.RangeDataRequest{
		FromTimestamp: time.Now().Add(-1 * time.Minute).UnixNano(),
		ToTimestamp:   time.Now().UnixNano(),
	}
	grpcServerMock := newMockTelematicsDataService_GetRangeDataServer()
	err = server.GetRangeData(ctx, rangeReq, grpcServerMock)
	assert.NoError(t, err)

	fmt.Printf("All data from GetRangeData: %v\n", grpcServerMock.responses)
}

type mockTelematicsDataService_GetRangeDataServer struct {
	protobuf.TelematicsDataService_GetRangeDataServer
	responses []*protobuf.TelematicsDataProto
}

func newMockTelematicsDataService_GetRangeDataServer() *mockTelematicsDataService_GetRangeDataServer {
	return &mockTelematicsDataService_GetRangeDataServer{
		responses: []*protobuf.TelematicsDataProto{},
	}
}

func (m *mockTelematicsDataService_GetRangeDataServer) Send(data *protobuf.TelematicsDataProto) error {
	m.responses = append(m.responses, data)
	return nil
}
