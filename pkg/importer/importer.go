package importer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/bf2fc6cc711aee1a0c2a/jira2sheets/pkg/config"
	"gopkg.in/errgo.v2/errors"
)

type Importer struct {
	Cfg                   *config.Config
	JiraPat               string
	GoogleCredentialsJson string
	Verbose               bool
}

type JiraCsvUrlParams struct {
	BaseUrl  string
	FilterId string
}

func (i *Importer) Run() error {
	log := log.Default()
	ctx := context.Background()
	if i.Verbose {
		log.Printf("config %+v", i.Cfg)
	}
	for _, spreadsheet := range i.Cfg.Spreadsheets {
		if i.Verbose {
			log.Printf("processing %s %s", spreadsheet.Url, spreadsheet.SheetName)
		}
		data := make([][]interface{}, 0)
		if i.Verbose {
			log.Printf("processing %s", spreadsheet.JiraFilter)
		}
		header, data, err := i.fetchCSVFromJIRA(spreadsheet.JiraFilter)
		if err != nil {
			return errors.Wrap(err)
		}
		if i.Verbose {
			log.Printf("loaded data from jira")
		}
		// Put all the CSV exports into a single data structure
		if len(data) > 0 {
			err := i.putCsvsToSheet(ctx, spreadsheet.Url, spreadsheet.SheetName, header, data)
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

	if i.Cfg.ActiveSprintsSheet.Url != "" && i.Cfg.ActiveSprintsSheet.SheetName != "" && i.Cfg.ActiveSprintsSheet.JiraEndpoint != "" {

		if i.Verbose {
			log.Printf("fetching active sprints using endopint: %s", i.Cfg.ActiveSprintsSheet.JiraEndpoint)
		}
		data, err := i.fetchApiGet(i.Cfg.ActiveSprintsSheet.JiraEndpoint)
		if err != nil {
			return errors.Wrap(err)
		}

		if len(data) > 0 {
			type Value struct {
				Id            int64  `json:"id"`
				Name          string `json:"name"`
				OriginBoardId int64  `json:"originBoardId"`
				StartDate     string `json:"startDate"`
				EndDate       string `json:"endDate"`
			}

			type Response struct {
				Values []Value `json:"values"`
			}

			var response Response
			json.Unmarshal([]byte(data), &response)

			if i.Verbose {
				log.Printf("active sprints: %+v", response.Values)
			}

			names := make([][]interface{}, len(response.Values))
			for i, v := range response.Values {
				board := make([]interface{}, 5)
				board[0] = v.Id
				board[1] = v.Name
				board[2] = v.OriginBoardId
				board[3] = v.StartDate
				board[4] = v.EndDate

				names[i] = board
			}

			headers := make([]interface{}, 5)
			headers[0] = "Sprint Id"
			headers[1] = "Name"
			headers[2] = "OriginBoardId"
			headers[3] = "Start Date"
			headers[4] = "End Date"

			err := i.putCsvsToSheet(ctx, i.Cfg.ActiveSprintsSheet.Url, i.Cfg.ActiveSprintsSheet.SheetName, headers, names)
			if err != nil {
				return errors.Wrap(err)
			}
		}
	}
	return nil
}
