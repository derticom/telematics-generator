package grpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"telematics-generator/pkg/cache"
	"telematics-generator/protobuf"
	"time"
)

type Server struct {
	cache *cache.TelematicsDataCache
	protobuf.UnimplementedTelematicsDataServiceServer
}

func NewServer(c *cache.TelematicsDataCache) *Server {
	return &Server{cache: c}
}

func (s *Server) GetLatestData(ctx context.Context, req *emptypb.Empty) (*protobuf.TelematicsDataProto, error) {
	data, ok := s.cache.GetLatest()
	if !ok {
		return nil, status.Error(codes.NotFound, "no data available")
	}

	return &protobuf.TelematicsDataProto{
		VehicleId: int32(data.VehicleID),
		Timestamp: data.Timestamp.UnixNano(),
		Speed:     int32(data.Speed),
		Latitude:  data.Latitude,
		Longitude: data.Longitude,
	}, nil
}

func (s *Server) GetRangeData(req *protobuf.RangeDataRequest, srv protobuf.TelematicsDataService_GetRangeDataServer) error {
	from := time.Unix(0, req.FromTimestamp)
	to := time.Unix(0, req.ToTimestamp)

	data, err := s.cache.GetRange(from, to)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	for _, d := range data {
		err := srv.Send(&protobuf.TelematicsDataProto{
			VehicleId: int32(d.VehicleID),
			Timestamp: d.Timestamp.UnixNano(),
			Speed:     int32(d.Speed),
			Latitude:  d.Latitude,
			Longitude: d.Longitude,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
