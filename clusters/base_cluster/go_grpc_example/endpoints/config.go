package endpoints

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// TODO: not sure where these should live, since the layer (db, controller, config)
// have not yet been separated.
const (
	ENV_DEV           = "DEV"
	ENV_SERV_HOST     = "HOST"
	ENV_SERV_PORT     = "PORT"
	SERV_HOST_DEFAULT = "127.0.0.1"
	SERV_PORT_DEFAULT = "80"
	HTTPS_CERT_PATH   = "/etc/secrets/host.cert"
	HTTPS_KEY_PATH    = "/etc/secrets/host.key"
)

type AppConfig struct {
	DbCreds DBCreds
	Addr    string
	Cert    string
	Key     string
}

func GetEnv(envVar, defaultVal string) string {
	viper.BindEnv(envVar)
	viper.SetDefault(envVar, defaultVal)
	return viper.GetString(envVar)
}

func GetTrimmedConfig(path, defaultCfg string) (string, error) {
	bytes, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return defaultCfg, nil
	}
	if err != nil {
		return "", fmt.Errorf("unable to read config file %s: %w", path, err)
	}

	return strings.TrimSpace(string(bytes)), nil
}

func ReadAppConfig() (*AppConfig, error) {
	dbCreds, err := ReadDBConfig()
	if err != nil {
		return nil, err
	}

	host := GetEnv(ENV_SERV_HOST, SERV_HOST_DEFAULT)
	port := GetEnv(ENV_SERV_PORT, SERV_PORT_DEFAULT)
	addr := fmt.Sprintf("%s:%s", host, port)

	// TODO: add encryption later. The mesh takes care of this, but it would be a useful exercise.
	cert := GetEnv(HTTPS_CERT_PATH, "")
	key := GetEnv(HTTPS_KEY_PATH, "")

	return &AppConfig{
		DbCreds: *dbCreds,
		Addr:    addr,
		Cert:    cert,
		Key:     key,
	}, nil
}
