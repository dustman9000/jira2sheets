package importer

import (
	"context"
	"fmt"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"gopkg.in/errgo.v2/errors"
	"log"
	"regexp"
)

var googleSpreadsheetMatcher = regexp.MustCompile("^https:\\/\\/docs\\.google\\.com\\/spreadsheets\\/d\\/(?P<spreadsheetId>.*)\\/edit.*$")

func (i *Importer) putCsvsToSheet(ctx context.Context, spreadsheetUrl string, sheetName string, header []interface{}, data [][]interface{}) error {
	srv, err := sheets.NewService(ctx, option.WithCredentialsJSON([]byte(i.GoogleCredentialsJson)), option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return errors.Because(err, nil, "Unable to retrieve Sheets client")
	}

	match := googleSpreadsheetMatcher.FindStringSubmatch(spreadsheetUrl)
	spreadsheetId := match[googleSpreadsheetMatcher.SubexpIndex("spreadsheetId")]
	updateRange := fmt.Sprintf("%s!1:%d", sheetName, len(data)+1)
	values := make([][]interface{}, 0)
	values = append(values, header)
	values = append(values, data...)
	updateValueRange := sheets.ValueRange{
		Values: values,
	}
	sheetId := int64(-1)
	existingColCount := int64(-1)
	existingRowCount := int64(-1)

	spreadsheet, err := srv.Spreadsheets.Get(spreadsheetId).Do()
	if err != nil {
		return errors.Wrap(err)
	}
	for _, s := range spreadsheet.Sheets {
		if s.Properties.Title == sheetName {
			sheetId = s.Properties.SheetId
			existingColCount = s.Properties.GridProperties.ColumnCount
			existingRowCount = s.Properties.GridProperties.RowCount
		}
	}
	spreadsheetRequest := sheets.BatchUpdateSpreadsheetRequest{
		Requests: make([]*sheets.Request, 0),
	}

	if existingRowCount > int64(len(data)+1) {
		spreadsheetRequest.Requests = append(spreadsheetRequest.Requests, &sheets.Request{
			DeleteDimension: &sheets.DeleteDimensionRequest{
				Range: &sheets.DimensionRange{
					Dimension:       "ROWS",
					SheetId:         sheetId,
					StartIndex:      int64(len(data)+1),
				},
			},
		})
	}
	if existingColCount > int64(len(header)) {
		spreadsheetRequest.Requests = append(spreadsheetRequest.Requests, &sheets.Request{
			DeleteDimension: &sheets.DeleteDimensionRequest{
				Range: &sheets.DimensionRange{
					Dimension:       "COLUMNS",
					SheetId:         sheetId,
					StartIndex:      int64(len(header)),
				},
			},
		})
	}

	if i.Verbose {
		log.Printf("clearing ranges %+v", spreadsheetRequest.Requests)
	}

	_, err = srv.Spreadsheets.Values.Update(spreadsheetId, updateRange, &updateValueRange).ValueInputOption("RAW").Do()
	if err != nil {
		return errors.Wrap(err)
	}
	if len(spreadsheetRequest.Requests) >0 {
		_, err = srv.Spreadsheets.BatchUpdate(spreadsheetId, &spreadsheetRequest).Do()
		if err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}

func intToLetters(number int64) (letters string) {
	number--
	if firstLetter := number / 26; firstLetter > 0 {
		letters += intToLetters(firstLetter)
		letters += string('A' + number%26)
	} else {
		letters += string('A' + number)
	}

	return
}
