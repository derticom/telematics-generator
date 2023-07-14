package main

import (
	"fmt"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"telematics-generator/pkg/cache"
	"telematics-generator/pkg/generator"
	mygrpc "telematics-generator/pkg/grpc"
	"telematics-generator/pkg/kafka"
	"telematics-generator/pkg/models"
	"telematics-generator/protobuf"
	"time"
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

func loadConfig(configPath string) (*AppConfig, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read the configuration file: %w", err)
	}

	vehiclesCountStr := viper.GetString("vehiclesCount")
	vehiclesCount, err := strconv.Atoi(vehiclesCountStr)
	if err != nil {
		return nil, fmt.Errorf("vehiclesCount should be an integer: %w", err)
	}
	if vehiclesCount > 100 {
		return nil, fmt.Errorf("vehiclesCount should be less than 100")
	}
	if vehiclesCount < 1 {
		return nil, fmt.Errorf("vehiclesCount should be more than 1")
	}

	maxSpeedStr := viper.GetString("maxSpeed")
	maxSpeed, err := strconv.Atoi(maxSpeedStr)
	if err != nil {
		return nil, fmt.Errorf("maxSpeed should be an integer: %w", err)
	}
	if maxSpeed < 1 {
		return nil, fmt.Errorf("maxSpeed should be more than 1")
	}
	if maxSpeed > 200 {
		return nil, fmt.Errorf("maxSpeed should be less than 200")
	}

	maxTimeStepStr := viper.GetString("maxTimeStep")
	maxTimeStep, err := time.ParseDuration(maxTimeStepStr)
	if err != nil {
		return nil, fmt.Errorf("invalid maxTimeStep format: %w", err)
	}
	if maxTimeStep > 24*time.Hour {
		return nil, fmt.Errorf("maxTimeStep should be less than 24h")
	}

	cacheSizeStr := viper.GetString("cacheSize")
	cacheSize, err := strconv.Atoi(cacheSizeStr)
	if err != nil {
		return nil, fmt.Errorf("cacheSize should be an integer: %w", err)
	}
	if cacheSize < 1 {
		return nil, fmt.Errorf("cacheSize should be more than 1")
	}
	if cacheSize > 1000_000 {
		return nil, fmt.Errorf("cacheSize should be less than 1 000 000")
	}

	brokerHost := viper.GetString("brokerHost")
	if brokerHost == "" {
		return nil, fmt.Errorf("brokerHost is required")
	}

	topicName := viper.GetString("topicName")
	if topicName == "" {
		return nil, fmt.Errorf("topicName is required")
	}

	grpsPortStr := viper.GetString("grpsPort")
	grpsPort, err := strconv.Atoi(grpsPortStr)
	if err != nil {
		return nil, fmt.Errorf("grpsPort should be an integer: %w", err)
	}
	if grpsPort < 0 {
		return nil, fmt.Errorf("grpsPort should be more than 0")
	}
	if grpsPort > 65536 {
		return nil, fmt.Errorf("grpsPort should be less than 65536")
	}

	return &AppConfig{
		VehiclesCount: vehiclesCount,
		MaxSpeed:      maxSpeed,
		MaxTimeStep:   int(maxTimeStep.Seconds()),
		CacheSize:     cacheSize,
		BrokerHost:    brokerHost,
		TopicName:     topicName,
		GrpsPort:      grpsPort,
	}, nil
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
