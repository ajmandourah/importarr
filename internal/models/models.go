package models

import "time"

type Instance struct {
	Name   string
	Type   string // "sonarr" or "radarr"
	URL    string
	APIKey string
}

type StatusMessage struct {
	Title    string   `json:"title"`
	Messages []string `json:"messages"`
}

type QueueRecord struct {
	ID                  int             `json:"id"`
	Title               string          `json:"title"`
	SeriesID            int             `json:"seriesId"`
	MovieID             int             `json:"movieId"`
	SeasonNumber        int             `json:"seasonNumber"`
	Protocol            string          `json:"protocol"`
	Status              string          `json:"status"`
	DownloadID          string          `json:"downloadId"`
	OutputPath          string          `json:"outputPath"`
	StatusMessages      []StatusMessage `json:"statusMessages"`
	EstimatedCompletion *time.Time      `json:"estimatedCompletionTime,omitempty"`
	Added               time.Time       `json:"added"`
}

func (r QueueRecord) SeriesOrMovieID() int {
	if r.SeriesID > 0 {
		return r.SeriesID
	}
	return r.MovieID
}

type ManualImportFile struct {
	Path               string   `json:"path,omitempty"`
	SeriesID           int      `json:"seriesId,omitempty"`
	MovieID            int      `json:"movieId,omitempty"`
	SeasonNumber       int      `json:"seasonNumber,omitempty"`
	EpisodeIDs         []int    `json:"episodeIds,omitempty"`
	DownloadID         string   `json:"downloadId,omitempty"`
	Quality            *Quality `json:"quality,omitempty"`
	Languages          []string `json:"languages,omitempty"`
	ReleaseGroup       string   `json:"releaseGroup,omitempty"`
	IndexerFlags       int      `json:"indexerFlags,omitempty"`
	ReleaseType        int      `json:"releaseType,omitempty"`
	Rejected           bool     `json:"rejected,omitempty"`
	PreviouslyImported bool     `json:"previouslyImported,omitempty"`
}

type Quality struct {
	Quality  QualityItem `json:"quality"`
	Revision Revision    `json:"revision"`
}

type QualityItem struct {
	Quality int    `json:"quality"`
	Name    string `json:"name"`
}

type Revision struct {
	Version int `json:"version"`
	Real    int `json:"real"`
}

type QueueResponse struct {
	Page         int           `json:"page"`
	PageSize     int           `json:"pageSize"`
	TotalRecords int           `json:"totalRecords"`
	Records      []QueueRecord `json:"records"`
}

type ImportResult struct {
	Path    string
	Status  string
	Message string
}
