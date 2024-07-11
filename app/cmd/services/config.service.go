package services

import (
	log "github.com/sirupsen/logrus"
	stdLog "log"
	"os"
	"path/filepath"
	"strings"

	configs "github.com/gulizay91/template-go-consumer/config"
	"github.com/spf13/viper"
)

var config *configs.Config

func InitConfig() {
	v := viper.New()

	var environment = os.Getenv("SERVICE__ENVIRONMENT")
	configName := "env"
	if environment != "" {
		configName = "env." + environment
	}
	stdLog.Printf("configName: %s", configName)
	//v.SetConfigType("dotenv")
	v.SetConfigType("yaml")
	v.SetConfigName(configName)
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	stdLog.Printf("workdir: %s", wd)
	stdLog.Printf("config path: %s", filepath.Dir(wd))
	//v.AddConfigPath("../")
	v.AddConfigPath(filepath.Dir(wd))
	v.AutomaticEnv()

	// used `__` nested config in .env files
	v.SetEnvKeyReplacer(strings.NewReplacer(`.`, `__`))

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := v.Unmarshal(&config); err != nil {
		panic(err)
	}

	if err := config.Validate(); err != nil {
		panic(err)
	}

	log.SetFormatter(&log.JSONFormatter{})
	logLevel := getLogLevel(config.Service.LogLevel)
	log.SetLevel(logLevel)
	log.Info("service logLevel: ", logLevel)
}

func getLogLevel(strLogLevel string) log.Level {
	logLevel := log.InfoLevel
	switch strLogLevel {
	case "trace":
		logLevel = log.TraceLevel
		break
	case "debug":
		logLevel = log.DebugLevel
		break
	case "warn":
		logLevel = log.WarnLevel
		break
	case "error":
		logLevel = log.ErrorLevel
		break
	case "fatal":
		logLevel = log.FatalLevel
		break
	case "panic":
		logLevel = log.PanicLevel
		break
	}
	return logLevel
}
