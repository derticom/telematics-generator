package main

import (
	"fmt"
	"generator/pkg/cache"
	"generator/pkg/generator"
	mygrpc "generator/pkg/grpc"
	"generator/pkg/kafka"
	"generator/pkg/models"
	"generator/protobuf"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type AppConfig struct {
	VehiclesCount int
	MaxSpeed      int
	MaxTimeStep   int
	CacheSize     int
	BrokerHost    string
	TopicName     string
	GrpsPort      int
}

func loadConfig(configPath string) (*AppConfig, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read the configuration file: %w", err)
	}

	return &AppConfig{
		VehiclesCount: viper.GetInt("vehiclesCount"),
		MaxSpeed:      viper.GetInt("maxSpeed"),
		MaxTimeStep:   viper.GetInt("maxTimeStep"),
		CacheSize:     viper.GetInt("cacheSize"),
		BrokerHost:    viper.GetString("brokerHost"),
		TopicName:     viper.GetString("topicName"),
		GrpsPort:      viper.GetInt("grpsPort"),
	}, nil
}

func main() {
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("Initializing Kafka producer")
	producer := kafka.NewKafkaProducer([]string{config.BrokerHost}, config.TopicName)

	log.Println("Initializing data generator")
	gen := generator.NewRandomTelematicsGenerator(config.MaxSpeed, config.MaxTimeStep)

	log.Println("Initializing data cache")
	telematicsDataCache := cache.NewTelematicsDataCache(config.CacheSize)

	log.Println("Initializing GRPC server")
	s := mygrpc.NewServer(telematicsDataCache)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GrpsPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	protobuf.RegisterTelematicsDataServiceServer(grpcServer, s)

	go func() {
		log.Println("Starting GRPC server")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	var wg sync.WaitGroup

	log.Println("Starting data generation")
	for i := 1; i < config.VehiclesCount+1; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			for telematicsData := range gen.Generate(id) {
				telematicsDataCache.Add(telematicsData)

				protoData := convertToProto(telematicsData)

				err = producer.ProduceMessage(protoData)
				if err != nil {
					log.Printf("Failed to produce message: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()
	log.Println("Data generation completed")

	err = producer.Close()
	if err != nil {
		log.Printf("Failed to close producer: %v", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	log.Println("Stopping GRPC server")
	grpcServer.GracefulStop()
}

func convertToProto(telematicsData models.TelematicsData) *protobuf.TelematicsDataProto {
	return &protobuf.TelematicsDataProto{
		VehicleId: int32(telematicsData.VehicleID),
		Timestamp: telematicsData.Timestamp.UnixNano(),
		Speed:     int32(telematicsData.Speed),
		Latitude:  telematicsData.Latitude,
		Longitude: telematicsData.Longitude,
	}
}
