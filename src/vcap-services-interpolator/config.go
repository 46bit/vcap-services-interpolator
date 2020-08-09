package main

import (
	"os"
	"strconv"
	"strings"

	"code.cloudfoundry.org/lager"
)

type Config struct {
	Logger     lager.Logger
	ListenPort uint

	ServerCertPath         string
	ServerKeyPath          string
	InstanceIdentityCAPath string
}

func NewConfigFromEnv() Config {
	return Config{
		Logger:     getDefaultLogger(),
		ListenPort: getEnvWithDefaultInt("PORT", 9299),

		ServerCertPath:         os.Getenv("SERVER_CERT_PATH"),
		ServerKeyPath:          os.Getenv("SERVER_KEY_PATH"),
		InstanceIdentityCAPath: os.Getenv("INSTANCE_IDENTITY_CA_PATH"),
	}
}

func getEnvWithDefaultInt(k string, def uint) uint {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	d, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		panic(err)
	}
	return uint(d)
}

func getDefaultLogger() lager.Logger {
	logger := lager.NewLogger("vcap-services-interpolator")
	logLevel := lager.INFO
	if strings.ToLower(os.Getenv("LOG_LEVEL")) == "debug" {
		logLevel = lager.DEBUG
	}
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, logLevel))

	return logger
}
