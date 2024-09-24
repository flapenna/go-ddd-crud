package main

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/flapenna/go-ddd-crud/config"
	domain "github.com/flapenna/go-ddd-crud/internal/domain/user"
	kafkaC "github.com/flapenna/go-ddd-crud/internal/infrastructure/kafka"
	"github.com/flapenna/go-ddd-crud/internal/infrastructure/mongodb"
	grpcServer "github.com/flapenna/go-ddd-crud/internal/interfaces/grpc"
	pbHealth "github.com/flapenna/go-ddd-crud/pkg/pb/health/v1"
	pb "github.com/flapenna/go-ddd-crud/pkg/pb/user/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()
	cfg := config.NewConfig()

	// Connect to MongoDB
	mongoConn := options.Client().ApplyURI(cfg.MongoDBUri)
	mongoClient, err := mongo.Connect(ctx, mongoConn)

	if err != nil {
		log.Fatal(err)
	}

	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Info("MongoDB successfully connected...")

	// MongoDb
	mongoDb := mongoClient.Database(cfg.MongoDBDatabase)

	// Drop the collection on every startup
	if err := mongoDb.Collection(cfg.MongoDBUserCollection, options.Collection().SetReadPreference(readpref.Secondary())).Drop(ctx); err != nil {
		log.Fatal(err)
	}

	// Create the collection with options (needed to return the pre-changes document using change stream)
	collOpts := options.CreateCollection().
		SetChangeStreamPreAndPostImages(bson.M{"enabled": true})

	err = mongoDb.CreateCollection(ctx, cfg.MongoDBUserCollection, collOpts)
	if err != nil {
		log.Fatalf("failed to set up an additional collection options: %v", err)
	}

	// Create new User Repository
	userCollection := mongoDb.Collection(cfg.MongoDBUserCollection)
	userRepo := mongodb.NewUserRepository(userCollection)

	// Kafka
	broker, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": cfg.KafkaServer})
	if err != nil {
		log.Fatalf("Failed to create producer due to %v", err)
	}
	defer broker.Close()

	userProducer := kafkaC.NewUserProducer(broker, "go-ddd-crud_user-event")

	// Create new user watcher
	userWatcher := mongodb.NewChangeStreamWatcher(userCollection)

	// Create user service
	userService := domain.NewUserService(userRepo, userProducer, userWatcher)

	// Set up gRPC server
	userServiceServer := grpcServer.NewUserServiceServer(userService)
	healthServiceServer := grpcServer.NewHealthServiceServer()

	grpcServer := grpc.NewServer()

	pb.RegisterUserServiceServer(grpcServer, userServiceServer)
	pbHealth.RegisterHealthServiceServer(grpcServer, healthServiceServer)

	// Enable reflection for the gRPC server (useful for debugging and testing)
	reflection.Register(grpcServer)

	// Start gRPC server
	go func() {
		log.Infof("Starting gRPC server on port %s", cfg.GrpcPort)
		lis, err := net.Listen("tcp", ":"+cfg.GrpcPort)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.NewClient(
		fmt.Sprintf("0.0.0.0:%s", cfg.GrpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwMux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
	}))

	// Register Health
	err = pbHealth.RegisterHealthServiceHandler(context.Background(), gwMux, conn)
	if err != nil {
		log.Fatalln("Failed to register Health handler gateway:", err)
	}

	// Register User
	err = pb.RegisterUserServiceHandler(context.Background(), gwMux, conn)
	if err != nil {
		log.Fatalln("Failed to register User handler to gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.HttpPort),
		Handler: gwMux,
	}

	go func() {
		log.Infof("Starting gRPC-Gateway on port %s", cfg.HttpPort)

		if err := gwServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to serve gateway: %v", err)
		}
	}()

	// Start user watcher
	userService.StartWatchingUsers(ctx)

	// Wait for interrupt signal to gracefully shut down the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-c
	defer signal.Stop(c)

	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	log.Println("gRPC server shut down")
}
