package importer

import (
	"bytes"
	"fmt"
	"gopkg.in/errgo.v2/errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"text/template"
)

var jiraFilterMatcher = regexp.MustCompile("^(?P<baseUrl>.*)\\/issues\\/\\?filter=(?P<filterId>\\d+)$")
var jiraCsvUrlTmpl = template.Must(template.New("JiraCsvUrl").Parse("{{ .BaseUrl }}/sr/jira.issueviews:searchrequest-csv-current-fields/{{ .FilterId }}/SearchRequest-{{ .FilterId }}.csv"))

func (i *Importer) fetchCSVFromJIRA(filterUrl string) ([][]interface{}, error) {
	log := log.Default()
	// parse the filter URL
	match := jiraFilterMatcher.FindStringSubmatch(filterUrl)
	var tpl bytes.Buffer
	if err := jiraCsvUrlTmpl.Execute(&tpl, JiraCsvUrlParams{
		BaseUrl:  match[jiraFilterMatcher.SubexpIndex("baseUrl")],
		FilterId: match[jiraFilterMatcher.SubexpIndex("filterId")],
	}); err != nil {
		return nil, errors.Wrap(err)
	}

	url := tpl.String()
	if i.Verbose {
		log.Printf("csv export url is %s", url)
	}
	pageSize := 500
	result := make([][]interface{}, 0)
	for p := 0; true; p++ {
		page, err := i.fetchCSVPageFromJIRA(url, p, pageSize)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		result = append(result, page...)
		if len(page) < pageSize {
			break
		}
	}
	return result, nil
}

func (i *Importer) fetchCSVPageFromJIRA(url string, page int, pageSize int) ([][]interface{}, error) {
	start := page * pageSize
	pagedUrl := fmt.Sprintf("%s?jqlQuery=&tempMax=%d&pager/start=%d", url, pageSize, start)
	req, err := http.NewRequest("GET", pagedUrl, nil)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	req.Header.Set("Content-Type", "text/csv;charset=UTF-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", i.JiraPat))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		result, err := parseCsv(resp.Body, true)
		if err != nil {

			return nil, errors.Because(err, nil, fmt.Sprintf("parsing csv from %s", url))
		}
		return result, nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return nil, fmt.Errorf("http error code %v %s %s", resp.StatusCode, resp.Status, string(bodyBytes))
}