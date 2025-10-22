package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/pflag"
	"gorm.io/gorm"

	"github.com/pachirode/distributed-task-system-demo/pkg/db"
)

const helpText = `Usage: main [flags] arg [arg...]

This is a pflag example.

Flags:
`

var (
	host      = pflag.String("host", "127.0.0.1:3306", "MySQL host address.")
	username  = pflag.StringP("username", "u", "root", "Username to connect to the database.")
	password  = pflag.StringP("password", "p", "system-watch", "Password to use when connecting to the database.")
	database  = pflag.StringP("db", "d", "system_watch", "Database name to connect to.")
	modelPath = pflag.String("model-pkg-path", "", "Generated model code's package name.")
	help      = pflag.BoolP("help", "h", false, "Show this help message.")

	usage = func() {
		fmt.Printf("%s", helpText)
		pflag.PrintDefaults()
	}
)

func main() {
	pflag.Usage = usage
	pflag.Parse()

	if *help {
		pflag.Usage()
		return
	}

	dbInstance, err := initializeDatabase()
	if err != nil {
		slog.Error("Failed to connect to database", "err", err)
	}

	modelPkgPath := resolveModelPackagePath(generateConfig.ModelPackagePath)

	generator := createGenerator(modelPkgPath)
	generator.UseDB(dbInstance)

	applyGeneratorOptions(generator)

	generateConfig.GenerateFunc(generator)

	generator.Execute()
}

func initializeDatabase() (*gorm.DB, error) {
	dbOptions := &db.MySQLOptions{
		Host:     *host,
		Username: *username,
		Password: *password,
		Database: *database,
	}

	return db.NewMySQL(dbOptions)
}
