package importer

import (
	"encoding/csv"
	"io"

	"gopkg.in/errgo.v2/errors"
)

func parseCsv(data io.Reader) ([]string, [][]string, error) {
	reader := csv.NewReader(data)
	reader.Comma = '|'
	reader.LazyQuotes = true
	csv, err := reader.ReadAll()
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	return csv[0], csv[1:], nil
}
