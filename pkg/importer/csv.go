package importer

import (
	"encoding/csv"
	"gopkg.in/errgo.v2/errors"
	"io"
)

func parseCsv(data io.Reader) ([]string, [][]string, error) {
	reader := csv.NewReader(data)
	csv, err := reader.ReadAll()
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	return csv[0], csv[1:], nil
}
