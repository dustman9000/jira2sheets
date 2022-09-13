package cmd

import (
	"log"
	"os"

	"github.com/bf2fc6cc711aee1a0c2a/jira2sheets/pkg/config"
	"github.com/bf2fc6cc711aee1a0c2a/jira2sheets/pkg/importer"
	"github.com/spf13/cobra"
	"gopkg.in/errgo.v2/errors"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Run one or more JIRA filters to a google sheet",
	Long:  ``,
	RunE:  run,
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringP("config", "c", "jira2sheets.yml", "Path to config file")
	importCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")
	importCmd.Flags().String("jira-pat", "", "The personal access token for accessing JIRA, normally read from JIRA2SHEETS_JIRA_PAT")
	importCmd.Flags().String("google-credentials-json", "", "The Google credentials.json, normally read from JIRA2SHEETS_GOOGLE_CREDENTIALS_JSON")
}

func run(cmd *cobra.Command, args []string) error {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return errors.Wrap(err)
	}
	configPath := cmd.Flag("config").Value.String()
	if verbose {
		log.Printf("config path %s", configPath)
	}
	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		return errors.Wrap(err)
	}
	jiraPat, err := cmd.Flags().GetString("jira-pat")
	if jiraPat == "" {
		jiraPat = os.Getenv("JIRA2SHEETS_JIRA_PAT")
	}
	if err != nil {
		return errors.Wrap(err)
	}
	googleCredentialsJson := cmd.Flag("google-credentials-json").Value.String()
	if googleCredentialsJson == "" {
		googleCredentialsJson = os.Getenv("JIRA2SHEETS_GOOGLE_CREDENTIALS_JSON")
	}

	importer := importer.Importer{
		Cfg:                   cfg,
		JiraPat:               jiraPat,
		GoogleCredentialsJson: googleCredentialsJson,
		Verbose:               verbose,
	}
	err = importer.Run()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}
