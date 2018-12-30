package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/exmonitor/exclient"
	"github.com/exmonitor/exclient/database"
	"github.com/exmonitor/exlogger"
	"github.com/exmonitor/firefly/service"
)

var Flags struct {
	// config file - not used atm
	ConfigFile string

	// logs
	LogToFile    bool
	LogFile      string
	LogErrorFile string

	// db
	DBDriver          string
	ElasticConnection string
	MariaConnection   string
	MariaUser         string
	MariaPassword     string

	// other
	TimeProfiling bool
	Debug         bool
}

var flags = Flags
var rootCmd = &cobra.Command{
	Use:   "carbon",
	Short: "carbon is a backend notification service for exmonitor system",
	Long: `Lotus is a backend notification service for exmonitor system.
Carbon fetches data from database and then run periodically  monitoring checks. 
Result of checks is stored back into database.
Every monitoring check run in separate thread to avoid delays because of IO operations.`,
}

func main() {

	// config file
	rootCmd.PersistentFlags().StringVarP(&flags.ConfigFile, "config", "c", "", "Set config file which will be used for fetching configuration.")

	// logs
	rootCmd.PersistentFlags().BoolVarP(&flags.LogToFile, "log-to-file", "", false, "Enable or disable logging to file.")
	rootCmd.PersistentFlags().StringVarP(&flags.LogFile, "log-file", "", "./notification.log", "Set filepath of log output. Used only when log-to-file is set to true.")
	rootCmd.PersistentFlags().StringVarP(&flags.LogErrorFile, "log-error-file", "", "./notification.error.log", "Set filepath of error log output. Used only when log-to-file is set to true.")

	// database
	rootCmd.PersistentFlags().StringVarP(&flags.DBDriver, "db-driver", "", "dummydb", "Set database driver that wil be used for connection")
	rootCmd.PersistentFlags().StringVarP(&flags.ElasticConnection, "elastic-connection", "", "", "Set elastic connection string.")
	rootCmd.PersistentFlags().StringVarP(&flags.MariaConnection, "maria-connection", "", "", "Set maria database connection string.")
	rootCmd.PersistentFlags().StringVarP(&flags.MariaUser, "maria-user", "", "", "Set Maria database user that wil be used for connection.")
	rootCmd.PersistentFlags().StringVarP(&flags.MariaPassword, "maria-password", "", "", "Set Maria database password that will be used for connection.")

	// other
	rootCmd.PersistentFlags().BoolVarP(&flags.Debug, "debug", "v", false, "Enable or disable more verbose log.")
	rootCmd.PersistentFlags().BoolVarP(&flags.TimeProfiling, "time-profiling", "", false, "Enable or disable time profiling.")

	rootCmd.Run = mainExecute

	err := rootCmd.Execute()

	if err != nil {
		panic(err)
	}
}

// main command execute function
func mainExecute(cmd *cobra.Command, args []string) {
	logConfig := exlogger.Config{
		Debug:        flags.Debug,
		LogToFile:    flags.LogToFile,
		LogFile:      flags.LogFile,
		LogErrorFile: flags.LogErrorFile,
	}

	logger, err := exlogger.New(logConfig)
	if err != nil {
		panic(err)
	}
	defer logger.CloseLogs()

	// database client connection
	var dbClient database.ClientInterface
	{
		// set db configuration
		dbConfig := exclient.DBConfig{
			DBDriver:          flags.DBDriver,
			ElasticConnection: flags.ElasticConnection,
			MariaConnection:   flags.MariaConnection,
			MariaUser:         flags.MariaConnection,
			MariaPassword:     flags.MariaPassword,
		}
		// init db client
		dbClient, err = exclient.GetDBClient(dbConfig)
		if err != nil {
			fmt.Printf("Failed to prepare DB Client.\n")
			panic(err)
		}
	}
	defer dbClient.Close()
	// catch Interrupt (Ctrl^C) or SIGTERM and exit
	// also make sure to close log files before exiting
	catchOSSignals(logger, dbClient)

	intervals, err := dbClient.SQL_GetIntervals()
	if err != nil {
		panic(err)
	}
	for _, interval := range intervals {
		// prepare main service/process config
		mainServiceConfig := service.Config{
			DBClient:      dbClient,
			FetchInterval: time.Duration(interval) * time.Second,

			Logger:        logger,
			TimeProfiling: flags.TimeProfiling,
		}
		// init main service/process
		mainService, err := service.New(mainServiceConfig)
		if err != nil {

		}
		// boot main service/process
		go mainService.Boot()
	}

	logger.Log("Main thread sleeping forever ....")
	select {}
}

// catch Interrupt (Ctrl^C) or SIGTERM and exit
func catchOSSignals(l *exlogger.Logger, dbClient database.ClientInterface) {
	// catch signals
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-c
		// be sure to close log files
		if flags.LogToFile {
			l.CloseLogs()
		}
		// clsoe db client
		dbClient.Close()

		fmt.Printf("\n>> Caught signal %s, exiting ...\n\n", s.String())
		os.Exit(1)
	}()
}
