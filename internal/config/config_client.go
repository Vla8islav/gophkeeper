// Package config parsing passed config values
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

// OptionsClient configuration parameters for the GophKeeper client.
//
// order of precedence: environment vars, command-line flags, config file, defaults.
type OptionsClient struct {
	ServerAddress OptionalString `env:"SERVER_ADDRESS" json:"server_address" command_arg:"base_url"`
	CACertPath    OptionalString `env:"CA_CERT" json:"ca_cert" command_arg:"ca_cert"`
	TokenFile     OptionalString `env:"TOKEN_FILE" json:"token_file" command_arg:"token_file"`
	Config        OptionalString `env:"CONFIG" json:"-" command_arg:"config"`
}

// ReadFlagsClient reads client configuration from command-line arguments and environment variables.
//
// Precedence: environment variables, command-line flags, config file, defaults.
func ReadFlagsClient(args []string, logger *zap.Logger) (*OptionsClient, []string, error) {
	if logger == nil {
		panic("config client logger is nil")
	}
	cmdOptions, rest, err := getOptionsClient(args)
	if err != nil {
		return nil, nil, fmt.Errorf("read command-line flags: %w", err)
	}
	logSetFlags(cmdOptions, logger)
	envOptions, err := getEnvOptionsClient()
	if err != nil {
		return nil, nil, fmt.Errorf("read environment variables: %w", err)
	}
	logSetEnv(envOptions, logger)
	var diskConfigOptions OptionsClient
	if cmdOptions.Config.BeenSet || envOptions.Config.BeenSet {
		var configFilename string
		if envOptions.Config.BeenSet && envOptions.Config.Value != "" {
			configFilename = envOptions.Config.Value
		} else if cmdOptions.Config.BeenSet && cmdOptions.Config.Value != "" {
			configFilename = cmdOptions.Config.Value
		}
		diskConfigOptions, err = getDiskConfigOptionsClient(configFilename)
		if err != nil {
			return nil, nil, fmt.Errorf("read config file: %w", err)
		}
		logConfigOptions(&diskConfigOptions, logger)
	}
	finalOptions := OptionsClient{
		ServerAddress: OptionalString{
			Value:   "https://localhost:8080",
			BeenSet: false,
		},
		CACertPath: OptionalString{
			Value:   "",
			BeenSet: false,
		},
		TokenFile: OptionalString{
			Value:   defaultClientTokenFile(),
			BeenSet: false,
		},
	}
	mergeOptions(&finalOptions, diskConfigOptions)
	mergeOptions(&finalOptions, *cmdOptions)
	mergeOptions(&finalOptions, *envOptions)
	return &finalOptions, rest, nil
}

func getDiskConfigOptionsClient(filename string) (OptionsClient, error) {
	if filename == "" {
		return OptionsClient{}, nil
	}
	configBytes, err := os.ReadFile(filename)
	if err != nil {
		return OptionsClient{}, err
	}
	options := OptionsClient{}
	if err = json.Unmarshal(configBytes, &options); err != nil {
		return OptionsClient{}, err
	}
	return options, nil
}

func getEnvOptionsClient() (*OptionsClient, error) {
	opt := OptionsClient{}
	if err := env.Parse(&opt); err != nil {
		return nil, err
	}
	return &opt, nil
}

func getOptionsClient(args []string) (*OptionsClient, []string, error) {
	opt := &OptionsClient{}
	fs := flag.NewFlagSet("gophkeeper-client", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Var(&opt.ServerAddress, "base-url", "базовый URL сервера, например https://localhost:8080")
	fs.Var(&opt.CACertPath, "custom-certificate", "путь до CA-сертификата для проверки сервера (необязательно)")
	fs.Var(&opt.TokenFile, "token", "путь до файла с токеном авторизации")
	fs.Var(&opt.Config, "config", "путь до файла с конфигурацией приложения")
	fs.Var(&opt.Config, "c", "путь до файла с конфигурацией приложения")
	if err := fs.Parse(args); err != nil {
		return nil, nil, err
	}
	return opt, fs.Args(), nil
}

func defaultClientTokenFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".gophkeeper_token"
	}
	return filepath.Join(home, ".gophkeeper", "token")
}
