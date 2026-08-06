package config

import (
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/caarlos0/env/v6"
)

// OptionsServer TODO: implement a clean option separation
type OptionsServer struct {
	ServerAddress          OptionalString `env:"RUN_ADDRESS"`
	AccrualAddress         OptionalString `env:"ACCRUAL_SYSTEM_ADDRESS"`
	AccrualPollingInterval OptionalInt    `env:"ACCRUAL_POLLING_INTERVAL"`

	DatabaseURI OptionalString `env:"DATABASE_URI"`

	MigrationsFolder OptionalString `env:"MIGRATIONS_FOLDER"`

	AuthTokenSecret OptionalString `env:"AUTH_TOKEN_SECRET"`
}

func logSetFlagsServer(options *OptionsServer) {
	if options == nil {
		return
	}
	var setFlags []string

	if options.ServerAddress.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-a=%s", options.ServerAddress.Value))
	}

	if options.AccrualAddress.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-r=%s", options.AccrualAddress.Value))
	}

	if options.DatabaseURI.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-d=%s", options.DatabaseURI.Value))
	}

	if options.MigrationsFolder.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-m=%s", options.MigrationsFolder.Value))
	}

	if options.AuthTokenSecret.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-s=%s", options.AuthTokenSecret.Value))
	}

	if options.AccrualPollingInterval.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-i=%d", options.AccrualPollingInterval.Value))
	}

	if len(setFlags) == 0 {
		log.Println("no command-line flags were set")
		return
	}

	for _, flagValue := range setFlags {
		log.Printf("command-line flag set: %s", flagValue)
	}
}

func logSetEnvServer(options *OptionsServer) {
	if options == nil {
		return
	}
	var setEnv []string

	if options.ServerAddress.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("ADDRESS=%s", options.ServerAddress.Value))
	}

	if options.AccrualAddress.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("ACCRUAL_ADDRESS=%s", options.AccrualAddress.Value))
	}

	if options.DatabaseURI.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("DATABASE_URI=%s", options.DatabaseURI.Value))
	}

	if options.MigrationsFolder.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("MIGRATIONS_FOLDER=%s", options.MigrationsFolder.Value))
	}

	if options.AuthTokenSecret.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("AUTH_TOKEN_SECRET=%s", options.AuthTokenSecret.Value))
	}

	if options.AccrualPollingInterval.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("ACCRUAL_POLLING_INTERVAL=%d", options.AccrualPollingInterval.Value))
	}

	if len(setEnv) == 0 {
		log.Println("no environment variables were set")
		return
	}

	for _, envValue := range setEnv {
		log.Printf("environment variable set: %s", envValue)
	}
}

func ReadFlagsServer(args []string) *OptionsServer {
	cmdOptions, err := getOptionsServer(args)
	if err != nil {
		log.Fatalln(err)
	}
	logSetFlagsServer(cmdOptions)

	envOptions := getEnvOptions()
	logSetEnvServer(envOptions)

	finalOptions := OptionsServer{
		ServerAddress:          OptionalString{Value: "localhost:8080", BeenSet: false},
		AccrualAddress:         OptionalString{Value: "", BeenSet: false},
		AccrualPollingInterval: OptionalInt{Value: 3, BeenSet: false},
		DatabaseURI: OptionalString{Value: "postgres://default_user:default_password@localhost:5432/gophkeeper_db?sslmode=disable",
			BeenSet: false},
		MigrationsFolder: OptionalString{Value: "./migrations", BeenSet: false},
		AuthTokenSecret:  OptionalString{Value: "super-duper-secret-dev-change-in-prod", BeenSet: false},
	}

	// env options are the priority
	mergeOptionsServer(&finalOptions, *cmdOptions)
	mergeOptionsServer(&finalOptions, *envOptions)

	//setOptionsTrue(&finalOptions)
	return &finalOptions
}

func mergeOptionsServer(mergeInto *OptionsServer, newValues OptionsServer) {
	if newValues.ServerAddress.BeenSet {
		mergeInto.ServerAddress = newValues.ServerAddress
		mergeInto.ServerAddress.BeenSet = true
	}

	if newValues.AccrualAddress.BeenSet {
		mergeInto.AccrualAddress = newValues.AccrualAddress
		mergeInto.AccrualAddress.BeenSet = true
	}

	if newValues.DatabaseURI.BeenSet {
		mergeInto.DatabaseURI = newValues.DatabaseURI
		mergeInto.DatabaseURI.BeenSet = true
	}

	if newValues.MigrationsFolder.BeenSet {
		mergeInto.MigrationsFolder = newValues.MigrationsFolder
		mergeInto.MigrationsFolder.BeenSet = true
	}

	if newValues.AuthTokenSecret.BeenSet {
		mergeInto.AuthTokenSecret = newValues.AuthTokenSecret
		mergeInto.AuthTokenSecret.BeenSet = true
	}

	if newValues.AccrualPollingInterval.BeenSet {
		mergeInto.AccrualPollingInterval = newValues.AccrualPollingInterval
		mergeInto.AccrualPollingInterval.BeenSet = true
	}

}

func getEnvOptions() *OptionsServer {
	var opt OptionsServer
	err := env.Parse(&opt)
	if err != nil {
		log.Fatalln(err)
	}
	return &opt
}

func getOptionsServer(args []string) (*OptionsServer, error) {

	opt := &OptionsServer{}

	fs := flag.NewFlagSet("metrics-aggregator-server", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // optional: silence flag errors in tests

	fs.Var(&opt.ServerAddress, "a", "адрес и порт запуска этого сервера")
	fs.Var(&opt.AccrualAddress, "r", "адрес и порт запуска сервера расчета баллов лояльности")
	fs.Var(&opt.AccrualAddress, "i", "частота поллинга сервера расчета баллов лояльности")

	fs.Var(&opt.DatabaseURI, "d", "connection string/dsn для postgres базы данных")

	fs.Var(&opt.MigrationsFolder, "m", "относительный путь до миграций, например ./migrations")
	fs.Var(&opt.AuthTokenSecret, "s", "секретный ключ для генерации токенов авторизации")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return opt, nil
}
