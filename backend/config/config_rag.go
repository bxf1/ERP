package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type LogConfig struct {
	Level  string
	Format string
}

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	Embedding EmbeddingConfig
	RAG       RAGConfig
	Log       LogConfig
}

type ServerConfig struct {
	Port int
	Mode string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

type JWTConfig struct {
	Secret string
	Expire int // hours
}

type EmbeddingConfig struct {
	Provider string // "openai" or "mock"
	APIKey   string
	BaseURL  string
	Model    string
}

type RAGConfig struct {
	TopK        int // number of document chunks to retrieve
	ChunkSize   int // target chunk size for document splitting
	ChunkOverlap int // overlap between chunks
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetEnvPrefix("ERP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("config file not found, using defaults and env: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	return &cfg
}

func setDefaults() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "erp")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("jwt.secret", "change-me-in-production")
	viper.SetDefault("jwt.expire", 24)
	viper.SetDefault("embedding.provider", "mock")
	viper.SetDefault("embedding.apikey", "")
	viper.SetDefault("embedding.baseurl", "https://api.openai.com/v1")
	viper.SetDefault("embedding.model", "text-embedding-ada-002")
	viper.SetDefault("rag.topk", 5)
	viper.SetDefault("rag.chunksize", 500)
	viper.SetDefault("rag.chunkoverlap", 50)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "console")
}
