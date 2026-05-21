package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SearchRequest sesuai dengan struktur JSON Pydantic di Python FastAPI
type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
	DOI   string `json:"doi,omitempty"`
}

type ChunkMetadata struct {
	ArticleID        string `json:"article_id"`
	Title            string `json:"title"`
	Authors          string `json:"authors"`
	DOI              string `json:"doi"`
	SectionHierarchy string `json:"section_hierarchy"`
}

// SearchResult menampung balikan dari Qdrant via Python API
type SearchResult struct {
	ChunkID  string        `json:"chunk_id"`
	Content  string        `json:"content"`
	Score    float64       `json:"score"`
	Metadata ChunkMetadata `json:"metadata"`
}

// SearchDatabase melakukan HTTP POST ke server Python RAG
func SearchDatabase(query string, doi string, limit int) (string, error) {
	if limit == 0 {
		limit = 5
	}

	reqBody := SearchRequest{
		Query: query,
		Limit: limit,
		DOI:   doi,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := "http://localhost:8000/search"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gagal memanggil RAG API (apakah uvicorn sudah menyala di port 8000?): %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("RAG API error (Kode HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var results []SearchResult
	if err := json.Unmarshal(bodyBytes, &results); err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "Tidak ada informasi yang relevan ditemukan di dalam database jurnal.", nil
	}

	// Format teks agar mudah dicerna oleh Gemini
	var finalContext string
	for i, r := range results {
		finalContext += fmt.Sprintf("--- Data %d (DOI: %s, Bab: %s) ---\n%s\n\n", i+1, r.Metadata.DOI, r.Metadata.SectionHierarchy, r.Content)
	}

	return finalContext, nil
}
