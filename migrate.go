package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	migrate "github.com/rubenv/sql-migrate"
)

type (
	flagvars struct {
		Help    bool
		Config  string
		Command string
		Number  int
	}
)

const (
	migrationFileTpl = "-- +migrate Up\n\n-- +migrate Down\n"
)

var (
	fv flagvars
)

// init initialize flags
func init() {
	const (
		shorthand = "see -%s for usage"

		helpUsage = "print this help"

		configDefault = "config.yml"
		configUsage   = "set config file"

		actionDefault = "up"
		actionUsage   = "migrate command, one of:\n\tup\n\tdown\n\tcreate\n"

		numberDefault = 0
		numberUsage   = "number of migrations\n Allow 0 for up as \"apply all\", required for down"
	)

	flag.BoolVar(&fv.Help, "h", false, fmt.Sprintf(shorthand, "help"))
	flag.BoolVar(&fv.Help, "help", false, helpUsage)

	flag.StringVar(&fv.Config, "c", configDefault, fmt.Sprintf(shorthand, "config"))
	flag.StringVar(&fv.Config, "config", configDefault, configUsage)

	flag.StringVar(&fv.Command, "a", actionDefault, fmt.Sprintf(shorthand, "action"))
	flag.StringVar(&fv.Command, "action", actionDefault, actionUsage)

	flag.IntVar(&fv.Number, "n", numberDefault, fmt.Sprintf(shorthand, "number"))
	flag.IntVar(&fv.Number, "number", numberDefault, numberUsage)
}

func printUsage() {
	fmt.Printf(`migrate [OPTION]... DIRECTORY
migrate -a create [OPTION]... DIRECTORY NEWNAME
wrapper for https://github.com/rubenv/sql-migrate with external migrations directory
DIRECTORY path to SQL migrations folder
NEWNAME name of new migration with create action
OPTION listed below
`)
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	if fv.Help {
		printUsage()
		os.Exit(0)
	}

	if !validate(fv) {
		fmt.Println("Wrong parameters passed:", strings.Join(os.Args[1:], " "))
		printUsage()
		os.Exit(1)
	}

	if flag.NArg() < 1 {
		fmt.Println("Migration path should passed. See -h for help")
		os.Exit(1)
	}

	if fv.Command == "down" && fv.Number == 0 {
		fmt.Println("Down migration required number of migrations. See -h for help")
		os.Exit(1)
	}

	if fv.Command == "create" && flag.NArg() < 2 {
		fmt.Println("Migration name required. See -h for help")
		os.Exit(1)
	}

	var err error

	var db *sql.DB

	if fv.Command != "create" {
		var conf Config
		if conf, err = NewConfig(fv.Config); err != nil {
			fmt.Printf("config init error: %s", err.Error())
			os.Exit(1)
		}

		db, err = conf.Postgres.Connect()
		if err != nil {
			fmt.Printf("db connect error: %s", err.Error())
			os.Exit(1)
		}
	}

	migrations := &migrate.FileMigrationSource{
		Dir: flag.Arg(0),
	}

	migrationDirection := migrate.Up
	if fv.Command == "down" {
		migrationDirection = migrate.Down
	}

	switch fv.Command {
	case "up", "down":
		applied, err := migrate.ExecMax(db, "postgres", migrations, migrationDirection, fv.Number)
		fmt.Printf("Rolling %s %d migrations!\n", fv.Command, applied)
		if err != nil {
			fmt.Println("Error occupied: ", err.Error())
			os.Exit(1)
		}
	case "create":
		filename := strconv.FormatInt(time.Now().Unix(), 10) + "_" + flag.Arg(1) + ".sql"
		path := filepath.Join(flag.Arg(0), filename)
		err := ioutil.WriteFile(path, []byte(migrationFileTpl), 0644)
		if err != nil {
			fmt.Println("Error occupied: ", err.Error())
			os.Exit(1)
		}
	}
}

func validate(fv flagvars) bool {
	var conditions = []bool{
		fv.Config != "",
		inSlice(fv.Command, []string{"up", "down", "create"}),
		fv.Number >= 0,
	}

	for _, c := range conditions {
		if !c {
			return false
		}
	}
	return true
}

func inSlice(search string, slice []string) bool {
	for _, s := range slice {
		if search == s {
			return true
		}
	}

	return false
}
