package config

import (
	"log"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port int
	Mode string
}

type DatabaseConfig struct {
	DSN         string
	MaxIdleConn int
	MaxOpenConn int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	ExpireTime int
}

type LogConfig struct {
	Level string
	Path  string
}

func LoadConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")

	// 设置环境变量
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Failed to read config file: %v, using defaults", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	// 从环境变量覆盖敏感配置
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		config.Database.DSN = dsn
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		config.JWT.Secret = secret
	}
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		config.Redis.Host = redisHost
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		config.Redis.Password = redisPassword
	}
	if serverPort := os.Getenv("SERVER_PORT"); serverPort != "" {
		if port, err := strconv.Atoi(serverPort); err == nil {
			config.Server.Port = port
		}
	}

	return &config
}

func setDefaults() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.dsn", "root:root123456@tcp(mysql:3306)/iam_db?charset=utf8mb4&parseTime=True&loc=Local")
	viper.SetDefault("database.maxIdleConn", 10)
	viper.SetDefault("database.maxOpenConn", 100)
	viper.SetDefault("redis.host", "redis")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("jwt.secret", "my-cloud-secret-key")
	viper.SetDefault("jwt.expireTime", 86400)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.path", "./logs")
}
