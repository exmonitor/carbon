package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

)

var Flags struct {
	// config file - not used atm
	ConfigFile string

	// db
	DBDriver   string
	DBServer   string
	DBUsername string
	DBPassword string
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
	// catch Interrupt (Ctrl^C) or SIGTERM and exit
	catchOSSignals()

	// set flags
	rootCmd.PersistentFlags().StringVarP(&flags.ConfigFile, "config", "c", "", "Set config file which will be used for fetching configuration.")

	rootCmd.PersistentFlags().StringVarP(&flags.DBDriver, "db-driver", "", "dummydb", "Set Database driver that wil be used for connection")
	rootCmd.PersistentFlags().StringVarP(&flags.DBServer, "db-server", "", "", "Set Database server that wil be used for connection")
	rootCmd.PersistentFlags().StringVarP(&flags.DBUsername, "db-username", "", "", "Set Database username that wil be used for connection")
	rootCmd.PersistentFlags().StringVarP(&flags.DBPassword, "db-password", "", "", "Set Database password that wil be used for connection")

	rootCmd.Run = mainExecute

	err := rootCmd.Execute()

	if err != nil {
		panic(err)
	}
}

// main command execute function
func mainExecute(cmd *cobra.Command, args []string) {
	dbClient, err := database.GetDBClient(flags.DBDriver, flags.DBServer, flags.DBUsername, flags.DBPassword)
	if err != nil {
		fmt.Printf("Failed to prepare DB Client.\n")
		panic(err)
	}



	// sleep little friend
	fmt.Printf(">> Main thread sleeping forever ...\n")
	select {}
}

// catch Interrupt (Ctrl^C) or SIGTERM and exit
func catchOSSignals() {
	// catch signals
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-c
		fmt.Printf("\n>> Caught signal %s, exiting ...\n\n", s.String())
		os.Exit(1)
	}()
}
