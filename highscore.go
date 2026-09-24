package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type highScoreStore struct {
	path string
}

type highScoreData struct {
	Score int `json:"score"`
}

func newHighScoreStore() (*highScoreStore, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dir = filepath.Join(dir, "snakr")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &highScoreStore{path: filepath.Join(dir, "highscore.json")}, nil
}

func (h *highScoreStore) load() int {
	data, err := os.ReadFile(h.path)
	if err != nil {
		return 0
	}
	var d highScoreData
	if err := json.Unmarshal(data, &d); err != nil {
		return 0
	}
	return d.Score
}

func (h *highScoreStore) save(score int) error {
	data, err := json.Marshal(highScoreData{Score: score})
	if err != nil {
		return err
	}
	return os.WriteFile(h.path, data, 0o644)
}
