package importer

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"text/template"

	"gopkg.in/errgo.v2/errors"
)

var jiraFilterMatcher = regexp.MustCompile("^(?P<baseUrl>.*)\\/issues\\/\\?filter=(?P<filterId>\\d+)$")

// https://issues.redhat.com/sr/jira.issueviews:searchrequest-csv-current-fields/12410471/SearchRequest-12410471.csv?delimiter=%7C
var jiraCsvUrlTmpl = template.Must(template.New("JiraCsvUrl").Parse("{{ .BaseUrl }}/sr/jira.issueviews:searchrequest-csv-current-fields/{{ .FilterId }}/SearchRequest-{{ .FilterId }}.csv?delimiter=%7C"))

func (i *Importer) fetchCSVFromJIRA(filterUrl string) ([]interface{}, [][]interface{}, error) {
	log := log.Default()
	// parse the filter URL
	match := jiraFilterMatcher.FindStringSubmatch(filterUrl)
	var tpl bytes.Buffer
	if err := jiraCsvUrlTmpl.Execute(&tpl, JiraCsvUrlParams{
		BaseUrl:  match[jiraFilterMatcher.SubexpIndex("baseUrl")],
		FilterId: match[jiraFilterMatcher.SubexpIndex("filterId")],
	}); err != nil {
		return nil, nil, errors.Wrap(err)
	}

	url := tpl.String()
	if i.Verbose {
		log.Printf("csv export url is %s", url)
	}
	pageSize := 999
	headerRows := make([][]string, 0)
	pages := make([][][]string, 0)
	for p := 0; true; p++ {
		headerRow, page, err := i.fetchCSVPageFromJIRA(url, p, pageSize)
		if err != nil {
			return nil, nil, errors.Wrap(err)
		}
		headerRows = append(headerRows, headerRow)
		pages = append(pages, page)
		if len(page) < pageSize {
			break
		}
	}
	headerResult, result := i.fixPadding(headerRows, pages)
	return headerResult, result, nil
}

func (i *Importer) fixPadding(headerRows [][]string, pages [][][]string) ([]interface{}, [][]interface{}) {
	// compute the col width
	maxColWidth := make(map[string]int)
	colWidths := make([]map[string]int, 0)
	colPaddings := make([][]int, 0)
	for _, headerRow := range headerRows {
		width := 0
		colWidth := make(map[string]int)
		for h, header := range headerRow {
			nextHeader := ""
			if (h + 1) < len(headerRow) {
				nextHeader = headerRow[h+1]
			} else {
				nextHeader = ""
			}
			if header != nextHeader {
				colWidth[header] = width
				if width >= maxColWidth[header] {
					maxColWidth[header] = width
				}
				width = 0
			} else {
				width++
			}
		}
		colWidths = append(colWidths, colWidth)
	}

	// create the header row
	headerRowResult := make([]interface{}, 0)
	// all header rows should have the same labels, just different widths
	for hr, headerRow := range headerRows {
		colPadding := make([]int, 0)
		for h, header := range headerRow {
			nextHeader := ""
			if (h + 1) < len(headerRow) {
				nextHeader = headerRow[h+1]
			} else {
				nextHeader = ""
			}
			if hr == 0 {
				headerRowResult = append(headerRowResult, header)
			}
			if header != nextHeader {
				colPadding = append(colPadding, maxColWidth[header]-colWidths[hr][header])
				if hr == 0 {
					// pad out the header row with more of the same labels
					for i := 0; i < colPadding[h]; i++ {
						headerRowResult = append(headerRowResult, header)
					}
				}
			} else {
				colPadding = append(colPadding, 0)
			}
		}
		colPaddings = append(colPaddings, colPadding)
	}

	// create the body
	result := make([][]interface{}, 0)
	for p, page := range pages {
		pageResult := make([][]interface{}, 0)
		for _, row := range page {
			rowResult := make([]interface{}, 0)
			for c, col := range row {
				rowResult = append(rowResult, col)
				for i := 0; i < colPaddings[p][c]; i++ {
					rowResult = append(rowResult, "")
				}
			}
			pageResult = append(pageResult, rowResult)
		}
		result = append(result, pageResult...)
	}
	return headerRowResult, result
}

func (i *Importer) fetchCSVPageFromJIRA(url string, page int, pageSize int) ([]string, [][]string, error) {
	start := page * pageSize
	pagedUrl := fmt.Sprintf("%s?jqlQuery=&tempMax=%d&pager/start=%d", url, pageSize, start)
	req, err := http.NewRequest("GET", pagedUrl, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	req.Header.Set("Content-Type", "text/csv;charset=UTF-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", i.JiraPat))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		headerRow, page, err := parseCsv(resp.Body)
		if err != nil {

			return nil, nil, errors.Because(err, nil, fmt.Sprintf("parsing csv from %s", url))
		}
		return headerRow, page, nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}

	return nil, nil, fmt.Errorf("http error code %v %s %s", resp.StatusCode, resp.Status, string(bodyBytes))
}

func (i *Importer) fetchApiGet(endpointUrl string) ([]byte, error) {

	req, err := http.NewRequest("GET", endpointUrl, nil)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", i.JiraPat))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	if resp.StatusCode == http.StatusOK {
		return bodyBytes, nil
	}

	return nil, fmt.Errorf("http error code %v %s %s", resp.StatusCode, resp.Status, string(bodyBytes))
}
