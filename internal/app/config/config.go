// Package config for getting service config variables from env, args and file.
package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"reflect"
)

// Struct contains all service settings.
//
// Settings read both from env and from args.
type Config struct {
	LocalURL      string `env:"SERVER_ADDRESS" json:"server_address"`
	Database      string `env:"DATABASE_DSN" json:"database_dsn"`
	S3Endpoint    string `env:"S3_ENDPOINT" json:"s3_endpoint"`
	S3AccessKey   string `env:"S3_ACCESS_KEY" json:"s3_access_key"`
	S3SecretKey   string `env:"S3_SECRET_KEY" json:"s3_secret_key"`
	S3Region      string `env:"S3_REGION" json:"s3_region"`
	S3Bucket      string `env:"S3_BUCKET" json:"s3_bucket"`
	SecretKey     string `env:"SECRET_KEY"`
	LogLevel      string `env:"LOG_LEVEL"`
	AuthTokenFile string `env:"AUTH_TOKEN_FILE"`
}

// Default config values.
var defaultConfig = Config{
	LocalURL:      "localhost:8080",
	Database:      "",
	S3Endpoint:    "localhost:9000",
	S3AccessKey:   "minioadmin",
	S3SecretKey:   "minioadmin",
	S3Region:      "localhost",
	S3Bucket:      "gopher",
	SecretKey:     "SECRET_KEY",
	LogLevel:      "info",
	AuthTokenFile: ".config",
}

// Parse command line flags.
func parseFlags(config *Config) {
	flag.StringVar(&config.LocalURL, "u", defaultConfig.LocalURL, "address and port to run server")
	flag.StringVar(&config.Database, "d", defaultConfig.Database, "database address")
	flag.StringVar(&config.S3Endpoint, "a", defaultConfig.S3Endpoint, "files s3 address")
	flag.StringVar(&config.S3AccessKey, "b", defaultConfig.S3AccessKey, "files s3 access key")
	flag.StringVar(&config.S3SecretKey, "c", defaultConfig.S3SecretKey, "files s3 secret key")
	flag.StringVar(&config.S3Region, "r", defaultConfig.S3Region, "files s3 region")
	flag.StringVar(&config.S3Bucket, "e", defaultConfig.S3Bucket, "files s3 bucket")
	flag.StringVar(&config.SecretKey, "k", defaultConfig.SecretKey, "secret key")
	flag.StringVar(&config.LogLevel, "l", defaultConfig.LogLevel, "log level")
	flag.StringVar(&config.AuthTokenFile, "t", defaultConfig.AuthTokenFile, "token file for server authorization")
	flag.Parse()
}

// Get config from env.
func updateFromEnv(config *Config) {
	v := reflect.Indirect(reflect.ValueOf(config))
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		var envName string
		if envName = field.Tag.Get("env"); envName == "" {
			continue
		}
		if envVal := os.Getenv(envName); envVal != "" {
			v.Field(i).SetString(envVal)
		}
	}
}

// Get config from file.
func updateDefaultFromConfigFile(configFile string) {
	if configFile == "" {
		return
	}
	file, err := os.ReadFile(configFile)
	if err != nil {
		log.Println("Cannot read config file")
		return
	}
	err = json.Unmarshal(file, &defaultConfig)
	if err != nil {
		log.Println("Wrong json config")
		return
	}
}

// Singleton variable.
var gophKeeperConfig *Config = nil

// Get overall config.
func GetConfig() Config {
	if gophKeeperConfig == nil {
		updateDefaultFromConfigFile(os.Getenv("CONFIG"))
		var config Config
		parseFlags(&config)
		updateFromEnv(&config)
		gophKeeperConfig = &config
	}
	return *gophKeeperConfig
}
