package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"importarr/internal/models"
)

type ArrClient interface {
	GetQueue() ([]models.QueueRecord, error)
	GetManualImport(record models.QueueRecord) ([]models.ManualImportFile, error)
	PostManualImport(files []models.ManualImportFile) ([]models.ImportResult, error)
	RemoveFromQueue(id int) error
	TriggerSearch(seriesOrMovieID, seasonNumber int) error
}

type baseClient struct {
	instance models.Instance
	client   *http.Client
}

func NewClient(inst models.Instance) (ArrClient, error) {
	b := &baseClient{
		instance: inst,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	switch inst.Type {
	case "sonarr":
		return &SonarrClient{baseClient: b}, nil
	case "radarr":
		return &RadarrClient{baseClient: b}, nil
	default:
		return nil, fmt.Errorf("unknown instance type: %s", inst.Type)
	}
}

func (b *baseClient) endpoint(path string) string {
	return b.instance.URL + path
}

func (b *baseClient) request(method, path string, body interface{}, resp interface{}) error {
	var req *http.Request
	var err error

	if body != nil {
		data, merr := json.Marshal(body)
		if merr != nil {
			return merr
		}
		req, err = http.NewRequest(method, b.endpoint(path), strings.NewReader(string(data)))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, b.endpoint(path), nil)
		if err != nil {
			return err
		}
	}

	req.Header.Set("X-Api-Key", b.instance.APIKey)
	req.Header.Set("Accept", "application/json")

	httpResp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode > 299 {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("API error %s: %s", httpResp.Status, string(bodyBytes))
	}

	if resp != nil {
		return json.NewDecoder(httpResp.Body).Decode(resp)
	}
	return nil
}

func DeduplicateByOutputPath(records []models.QueueRecord) []models.QueueRecord {
	seen := make(map[string]bool)
	var deduped []models.QueueRecord
	for _, r := range records {
		if !seen[r.OutputPath] {
			seen[r.OutputPath] = true
			deduped = append(deduped, r)
		}
	}
	return deduped
}

func isStuck(record models.QueueRecord) bool {
	for _, sm := range record.StatusMessages {
		for _, msg := range sm.Messages {
			if strings.Contains(strings.ToLower(msg), "via grab history") {
				return true
			}
		}
	}
	return false
}

func marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func jsonNewDecoder(r io.Reader) *json.Decoder {
	return json.NewDecoder(r)
}
