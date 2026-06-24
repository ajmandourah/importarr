package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"importarr/internal/models"
)

type RadarrClient struct {
	*baseClient
}

func (c *RadarrClient) GetQueue() ([]models.QueueRecord, error) {
	var resp models.QueueResponse
	err := c.request("GET", "/api/v3/queue?pageSize=5000", nil, &resp)
	if err != nil {
		return nil, err
	}

	var records []models.QueueRecord
	for _, r := range resp.Records {
		if isStuck(r) {
			records = append(records, r)
		}
	}
	return records, nil
}

func (c *RadarrClient) GetManualImport(record models.QueueRecord) ([]models.ManualImportFile, error) {
	params := url.Values{}
	params.Add("folder", record.OutputPath)
	params.Add("downloadId", record.DownloadID)
	params.Add("filterExistingFiles", "true")

	var files []models.ManualImportFile
	err := c.request("GET", "/api/v3/manualimport?"+params.Encode(), nil, &files)
	if err != nil {
		return nil, err
	}

	for i := range files {
		files[i].MovieID = record.SeriesOrMovieID()
		files[i].DownloadID = record.DownloadID
		files[i].FolderName = filepath.Base(record.OutputPath)
	}
	return files, nil
}

func (c *RadarrClient) PostManualImport(files []models.ManualImportFile) ([]models.ImportResult, error) {
	type postFile struct {
		Path         string            `json:"path"`
		FolderName   string            `json:"folderName"`
		MovieID      int               `json:"movieId"`
		Quality      *models.Quality   `json:"quality"`
		Languages    []models.Language `json:"languages"`
		ReleaseGroup string            `json:"releaseGroup"`
		IndexerFlags int               `json:"indexerFlags"`
		DownloadID   string            `json:"downloadId"`
	}

	postFiles := make([]postFile, len(files))
	for i, f := range files {
		postFiles[i] = postFile{
			Path:         f.Path,
			FolderName:   f.FolderName,
			MovieID:      f.MovieID,
			Quality:      f.Quality,
			Languages:    f.Languages,
			ReleaseGroup: f.ReleaseGroup,
			IndexerFlags: f.IndexerFlags,
			DownloadID:   f.DownloadID,
		}
	}

	jsonData, err := json.Marshal(postFiles)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseClient.endpoint("/api/v3/manualimport"), strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", c.instance.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("import failed: %s", resp.Status)
	}

	var results []postFile
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	var importResults []models.ImportResult
	for _, r := range results {
		status := "imported"
		message := ""
		importResults = append(importResults, models.ImportResult{
			Path:    r.Path,
			Status:  status,
			Message: message,
		})
	}
	return importResults, nil
}

func (c *RadarrClient) RemoveFromQueue(id int) error {
	return c.request("DELETE", fmt.Sprintf("/api/v3/queue/%d?removeFromClient=true", id), nil, nil)
}

func (c *RadarrClient) TriggerSearch(seriesOrMovieID, seasonNumber int) error {
	type SearchCommand struct {
		Name    string `json:"name"`
		MovieID int    `json:"movieId"`
	}
	cmd := SearchCommand{
		Name:    "moviesSearch",
		MovieID: seriesOrMovieID,
	}
	_ = c.request("POST", "/api/v3/command", cmd, nil)
	return nil
}
