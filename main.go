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
	"github.com/exmonitor/firefly/notification/email"
	"github.com/exmonitor/firefly/service"
	"gopkg.in/gomail.v2"
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
	MariaDatabaseName string
	MariaUser         string
	MariaPassword     string
	CacheEnabled      bool
	CacheTTl          string

	// smtp
	SMTPEnabled  bool
	EmailFrom    string
	SMTPServer   string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string

	// other
	TimeProfiling bool
	Debug         bool
}

var flags = Flags
var rootCmd = &cobra.Command{
	Use:   "firefly",
	Short: "firefly is a backend notification service for exmonitor system",
	Long:  `Firefly is a backend notification service for exmonitor system.`,
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
	rootCmd.PersistentFlags().StringVarP(&flags.ElasticConnection, "elastic-connection", "", "http://127.0.0.1:9200", "Set elastic connection string.")
	rootCmd.PersistentFlags().StringVarP(&flags.MariaConnection, "maria-connection", "", "", "Set maria database connection string.")
	rootCmd.PersistentFlags().StringVarP(&flags.MariaDatabaseName, "maria-database-name", "", "", "Set maria database name.")
	rootCmd.PersistentFlags().StringVarP(&flags.MariaUser, "maria-user", "", "", "Set Maria database user that wil be used for connection.")
	rootCmd.PersistentFlags().StringVarP(&flags.MariaPassword, "maria-password", "", "", "Set Maria database password that will be used for connection.")

	// smtp
	rootCmd.PersistentFlags().BoolVarP(&flags.SMTPEnabled, "smtp", "", true, "Enable or disable using the real SMTP server. If false, sending email is mocked.")
	rootCmd.PersistentFlags().StringVarP(&flags.EmailFrom, "smtp-email-from", "", "alert@alertea.com", "Default email that will be used in 'From'.")
	rootCmd.PersistentFlags().StringVarP(&flags.SMTPServer, "smtp-server", "", "127.0.0.1", "Hostname for SMTP server.")
	rootCmd.PersistentFlags().IntVarP(&flags.SMTPPort, "smtp-port", "", 0, "Port for SMTP server.")
	rootCmd.PersistentFlags().StringVarP(&flags.SMTPUser, "smtp-user", "", "alert@alertea.com", "Username for SMTP server.")
	rootCmd.PersistentFlags().StringVarP(&flags.SMTPPassword, "smtp-passwrd", "", "", "Password for SMTP server.")

	// cache
	rootCmd.PersistentFlags().BoolVarP(&flags.CacheEnabled, "cache", "", false, "Enable or disable caching of db records")
	rootCmd.PersistentFlags().StringVarP(&flags.CacheTTl, "cache-ttl", "", "5m", "Set cache ttl. Must be in time.Duration format. Value lower than 1m doesnt make sense.")

	// other
	rootCmd.PersistentFlags().BoolVarP(&flags.Debug, "debug", "v", false, "Enable or disable more verbose log.")
	rootCmd.PersistentFlags().BoolVarP(&flags.TimeProfiling, "time-profiling", "", false, "Enable or disable time profiling.")

	rootCmd.Run = mainExecute

	err := rootCmd.Execute()

	if err != nil {
		panic(err)
	}
}

func validateFlags() {
	if flags.TimeProfiling && !flags.Debug {
		fmt.Printf("WARNING: time profiling is shown via debug log, if you dont enabled debug log you wont see time profiling output.\n")
	}
}

// main command execute function
func mainExecute(cmd *cobra.Command, args []string) {
	validateFlags()

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
		// parse cache ttl
		cacheTTL, err := time.ParseDuration(flags.CacheTTl)
		if err != nil {
			fmt.Printf("Failed to parse cache TTL. %s is not valid format for time.Duration\n", flags.CacheTTl)
			panic(err)
		}

		// set db configuration
		dbConfig := exclient.DBConfig{
			DBDriver:          flags.DBDriver,
			ElasticConnection: flags.ElasticConnection,
			MariaConnection:   flags.MariaConnection,
			MariaDatabaseName: flags.MariaDatabaseName,
			MariaUser:         flags.MariaUser,
			MariaPassword:     flags.MariaPassword,

			CacheEnabled: flags.CacheEnabled,
			CacheTTL:     cacheTTL,

			Logger:        logger,
			TimeProfiling: flags.TimeProfiling,
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

	var emailChan chan *gomail.Message
	// email section
	if flags.SMTPEnabled {
		// email channel
		emailChan = email.BuildEmailChannel()
		// start email daemon
		emailDaemonConfig := email.DaemonConfig{
			SMTPConfig: email.SMTPConfig{
				Server:   flags.SMTPServer,
				Port:     flags.SMTPPort,
				Username: flags.SMTPUser,
				Password: flags.SMTPPassword,
				SMTPFrom: flags.EmailFrom,
			},
			Logger:    logger,
			EmailChan: emailChan,
		}

		emailDaemon, err := email.NewDaemon(emailDaemonConfig)
		if err != nil {
			fmt.Printf("Failed to start email daemon.\n")
			panic(err)
		}
		// run daemon
		emailDaemon.StartDaemon()
	}
	// make sure to close channel
	defer func() {
		if flags.SMTPEnabled {
			close(emailChan)
		}
	}()

	intervals, err := dbClient.SQL_GetIntervals()
	if err != nil {
		panic(err)
	}
	for _, interval := range intervals {
		// prepare main service/process config
		mainServiceConfig := service.Config{
			DBClient:      dbClient,
			FetchInterval: time.Duration(interval) * time.Second,
			SMTPEnabled:flags.SMTPEnabled,

			Logger:        logger,
			TimeProfiling: flags.TimeProfiling,
		}
		// init main service/process
		mainService, err := service.New(mainServiceConfig)
		if err != nil {
			panic(err)
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
		// close db client
		dbClient.Close()

		fmt.Printf("\n>> Caught signal %s, exiting ...\n\n", s.String())
		os.Exit(1)
	}()
}
