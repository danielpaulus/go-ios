package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func exitIfError(msg string, err error) {
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatalf(msg)
	}
}

var rootCmd = &cobra.Command{
	Use:   "go-ios",
	Short: "Cross-platform, open-source solution to interact with iOS devices",
	Long: `
go-ios is a cross-platform, open-source solution to interact with iOS devices.
It allows to run UI tests, launch or kill apps, install apps, pair devices, and more.
	`,
}
