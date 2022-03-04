package importer

import (
	"encoding/csv"
	"gopkg.in/errgo.v2/errors"
	"io"
)

func parseCsv(data io.Reader, removeHeaderRow bool) ([][]interface{}, error) {
	reader := csv.NewReader(data)
	csv, err := reader.ReadAll()
	if err != nil {
		return nil, errors.Wrap(err)
	}
	result := make([][]interface{}, 0)
	for i, row := range csv {
		if removeHeaderRow && i == 0 {
			continue
		}
		values := make([]interface{}, 0)
		for _, val := range row {
			values = append(values, val)
		}
		result = append(result, values)
	}
	return result, nil
}
