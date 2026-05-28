package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/user/lector/internal/plugin"
)

type RemoteMetadata struct {
	CoverURL string
	Synopsis string
}

func FetchRemoteMetadata(title, author string) *RemoteMetadata {
	cleanTitle := title
	delimiters := []string{"[", "(", "{", " - "}
	for _, char := range delimiters {
		if idx := strings.Index(title, char); idx > 0 {
			cleanTitle = strings.TrimSpace(title[:idx])
			break
		}
	}

	fetch := func(q string) *RemoteMetadata {
		apiURL := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?q=%s&maxResults=1", url.QueryEscape(q))
		req, _ := http.NewRequest("GET", apiURL, nil)

		resp, err := plugin.HTTPClient.Do(req)

		if err != nil || resp.StatusCode != 200 {
			return nil
		}
		defer resp.Body.Close()

		var result struct {
			Items []struct {
				VolumeInfo struct {
					Title       string `json:"title"`
					Description string `json:"description"`
					ImageLinks  struct {
						Thumbnail string `json:"thumbnail"`
					} `json:"imageLinks"`
				} `json:"volumeInfo"`
			} `json:"items"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || len(result.Items) == 0 {
			return nil
		}

		item := result.Items[0].VolumeInfo
		meta := &RemoteMetadata{Synopsis: item.Description}
		thumb := item.ImageLinks.Thumbnail
		if thumb != "" {
			meta.CoverURL = strings.ReplaceAll(strings.ReplaceAll(thumb, "http://", "https://"), "&edge=curl", "")
		}
		return meta
	}

	query := "intitle:" + cleanTitle
	if author != "" && !strings.EqualFold(author, "unknown") {
		query += "+inauthor:" + author
	}

	res := fetch(query)
	if res == nil || res.CoverURL == "" {
		res = fetch(cleanTitle)
	}
	return res
}
