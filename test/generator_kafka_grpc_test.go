package test

import (
	"context"
	"testing"
	"time"

	"generator/pkg/cache"
	"generator/pkg/generator"
	"generator/pkg/grpc"
	"generator/pkg/kafka"
	"generator/protobuf"
	"github.com/stretchr/testify/assert"
)

func TestGeneratorKafkaGRPC(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaProducer := kafka.NewKafkaProducer([]string{"localhost:9092"}, "topic1")

	gen := generator.NewRandomTelematicsGenerator(180, 5)

	dataCache := cache.NewTelematicsDataCache(100)

	go func() {
		dataCh := gen.Generate(1)
		for data := range dataCh {
			err := kafkaProducer.ProduceMessage(&protobuf.TelematicsDataProto{
				VehicleId: int32(data.VehicleID),
				Timestamp: data.Timestamp.UnixNano(),
				Speed:     int32(data.Speed),
				Latitude:  data.Latitude,
				Longitude: data.Longitude,
			})
			assert.NoError(t, err)

			dataCache.Add(data)

		}
	}()

	time.Sleep(10 * time.Second)

	server := grpc.NewServer(dataCache)

	data, err := server.GetLatestData(ctx, &protobuf.LatestDataRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, data)

	rangeReq := &protobuf.RangeDataRequest{
		FromTimestamp: time.Now().Add(-1 * time.Minute).UnixNano(),
		ToTimestamp:   time.Now().UnixNano(),
	}
	grpcServerMock := newMockTelematicsDataService_GetRangeDataServer()
	err = server.GetRangeData(rangeReq, grpcServerMock)
	assert.NoError(t, err)
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
