package config

import "os"

// Config is the Service's configuration object
type Config struct {
	GrpcPort              string
	HttpPort              string
	MongoDBUri            string
	MongoDBDatabase       string
	MongoDBUserCollection string
	KafkaServer           string
}

func NewConfig() *Config {
	return &Config{
		GrpcPort:              getEnv("SERVICE_GRPC_PORT", "8080"),
		HttpPort:              getEnv("SERVICE_HTTP_PORT", "8090"),
		MongoDBUri:            getEnv("MONGODB_URI", "mongodb://root:root@localhost:27017/"),
		MongoDBDatabase:       getEnv("MONGODB_DB", "go-ddd-crud"),
		MongoDBUserCollection: getEnv("MONGODB_USER_COLLECTION", "users"),
		KafkaServer:           getEnv("KAFKA_SERVER", "localhost:9092"),
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
