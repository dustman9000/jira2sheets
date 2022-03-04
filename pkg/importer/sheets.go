package importer

import (
	"context"
	"fmt"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"gopkg.in/errgo.v2/errors"
	"regexp"
)

var googleSpreadsheetMatcher = regexp.MustCompile("^https:\\/\\/docs\\.google\\.com\\/spreadsheets\\/d\\/(?P<spreadsheetId>.*)\\/edit.*$")

func (i *Importer) putCsvsToSheet(ctx context.Context, spreadsheetUrl string, sheetName string, data [][]interface{}) error {
	srv, err := sheets.NewService(ctx, option.WithCredentialsJSON([]byte(i.GoogleCredentialsJson)), option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return errors.Because(err, nil, "Unable to retrieve Sheets client")
	}

	match := googleSpreadsheetMatcher.FindStringSubmatch(spreadsheetUrl)
	spreadsheetId := match[googleSpreadsheetMatcher.SubexpIndex("spreadsheetId")]
	spreadsheetRange := fmt.Sprintf("%s!1:100000", sheetName);

	valueRange := sheets.ValueRange{
		Values: data,
			}
	_, err = srv.Spreadsheets.Values.Clear(spreadsheetId, spreadsheetRange, &sheets.ClearValuesRequest{}).Do()
	if err != nil {
		return errors.Wrap(err)
	}
	_, err = srv.Spreadsheets.Values.Update(spreadsheetId, spreadsheetRange, &valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}
