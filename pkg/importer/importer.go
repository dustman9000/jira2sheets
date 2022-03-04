package importer

import (
	"context"
	"github.com/pmuir/jira2sheets/pkg/config"
	"gopkg.in/errgo.v2/errors"
	"log"
)

type Importer struct {
	Cfg     *config.Config
	JiraPat string
	GoogleCredentialsJson string
	Verbose bool
}

type JiraCsvUrlParams struct {
	BaseUrl  string
	FilterId string
}

func (i *Importer) Run() error {
	log := log.Default()
	ctx := context.Background()
	if i.Verbose {
		log.Printf("config %v", i.Cfg)
	}
	for _, spreadsheet := range i.Cfg.Spreadsheets {
		if i.Verbose {
			log.Printf("processing %s %s", spreadsheet.Url, spreadsheet.SheetName)
		}
		data := make([][]interface{}, 0)
		for _, filterUrl := range spreadsheet.JiraFilters {
			if i.Verbose {
				log.Printf("processing %s", filterUrl)
			}
			csv, err := i.fetchCSVFromJIRA(filterUrl)
			if err != nil {
				return errors.Wrap(err)
			}
			if i.Verbose {
				log.Printf("loaded data from jira")
			}
			// Put all the CSV exports into a single data structure
			data = append(data, csv...)
		}
		if len(data) > 0 {
			err := i.putCsvsToSheet(ctx, spreadsheet.Url, spreadsheet.SheetName, data)
			if err != nil {
				return errors.Wrap(err)
			}
		} else {
			log.Printf("no data loaded")
		}
		if i.Verbose {
			log.Printf("finished processing %s %s", spreadsheet.Url, spreadsheet.SheetName)
		}
	}
	return nil
}
