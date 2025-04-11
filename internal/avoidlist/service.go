package avoidlist

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"wallet-guesser/internal/domain"

	log "github.com/sirupsen/logrus"
)

const (
	// DefaultAvoidListPath is the default path to the avoid list file
	DefaultAvoidListPath = "data/avoidlist.json"
)

// Service implements the AvoidListService interface
type Service struct {
	apiEndpoint string
	apiKey      string
	filePath    string
	entries     map[string]domain.AvoidListEntry
	lastUpdated time.Time
	mutex       sync.RWMutex
}

// NewService creates a new AvoidListService
func NewService(apiKey string, filePath string) *Service {
	if filePath == "" {
		filePath = DefaultAvoidListPath
	}

	return &Service{
		apiEndpoint: "https://api.dune.com/api/v1/query/4966121/results",
		apiKey:      apiKey,
		filePath:    filePath,
		entries:     make(map[string]domain.AvoidListEntry),
	}
}

// LoadFromFile loads the avoid list from a file
func (s *Service) LoadFromFile() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if file exists
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		log.Infof("Avoid list file not found: %s", s.filePath)
		return nil
	}

	file, err := os.Open(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to open avoid list file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read avoid list file: %w", err)
	}

	var fileData struct {
		Entries     []domain.AvoidListEntry `json:"entries"`
		LastUpdated time.Time               `json:"lastUpdated"`
	}

	if err := json.Unmarshal(data, &fileData); err != nil {
		return fmt.Errorf("failed to unmarshal avoid list data: %w", err)
	}

	// Populate the in-memory map
	s.entries = make(map[string]domain.AvoidListEntry, len(fileData.Entries))
	for _, entry := range fileData.Entries {
		s.entries[entry.Prefix] = entry
	}
	s.lastUpdated = fileData.LastUpdated

	log.Infof("Loaded %d avoid list entries, last updated at %s", len(s.entries), s.lastUpdated.Format(time.RFC3339))
	return nil
}

// SaveToFile saves the avoid list to a file
func (s *Service) saveToFile() error {

	// Create directory if it doesn't exist
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for avoid list: %w", err)
	}

	// Convert map to slice for JSON serialization
	entries := make([]domain.AvoidListEntry, 0, len(s.entries))
	for _, entry := range s.entries {
		entries = append(entries, entry)
	}

	fileData := struct {
		Entries     []domain.AvoidListEntry `json:"entries"`
		LastUpdated time.Time               `json:"lastUpdated"`
	}{
		Entries:     entries,
		LastUpdated: s.lastUpdated,
	}

	data, err := json.Marshal(fileData)
	if err != nil {
		return fmt.Errorf("failed to marshal avoid list data: %w", err)
	}

	// Write to file
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write avoid list file: %w", err)
	}

	log.Infof("Saved %d avoid list entries to %s", len(s.entries), s.filePath)
	return nil
}

// UpdateAvoidList updates the avoid list from the remote API
func (s *Service) UpdateAvoidList() error {
	if s.apiKey == "" {
		return fmt.Errorf("API key not provided for avoid list")
	}

	// if updated in past day, warn the user
	if s.lastUpdated.After(time.Now().Add(-time.Hour * 24)) {
		log.Warnf("Avoid list is already updated within the past day: %s", s.lastUpdated)
		return nil
	}

	url := fmt.Sprintf("%s?api_key=%s", s.apiEndpoint, s.apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to request avoid list data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read API response: %w", err)
	}

	var response domain.AvoidListResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return fmt.Errorf("failed to unmarshal API response: %w", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Update the entries map
	s.entries = make(map[string]domain.AvoidListEntry, len(response.Result.Rows))
	for _, entry := range response.Result.Rows {
		for idx, _ := range entry.ZippedPrefix {
			prefix := entry.ZippedPrefix[idx]
			_type := entry.ZippedType[idx]
			s.entries[prefix] = domain.AvoidListEntry{
				Prefix: prefix,
				Type:   _type,
			}
		}
	}
	s.lastUpdated = time.Now()

	// Save to file
	if err := s.saveToFile(); err != nil {
		log.Errorf("Failed to save avoid list: %v", err)
		// Continue anyway as we've updated the in-memory list
	}

	log.Infof("Updated avoid list with %d entries", len(s.entries))
	return nil
}

// ShouldAvoid checks if an address should be avoided
func (s *Service) ShouldAvoid(address string) (bool, string) {
	// Need at least 8 characters to check
	if len(address) < 8 {
		return false, ""
	}

	prefix := address[:8]
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if entry, ok := s.entries[prefix]; ok {
		if entry.Type == "t" {
			return true, "token with too many holders (>100k)"
		} else if entry.Type == "w" {
			return true, "wallet with too many tokens (>500)"
		}
	}

	return false, ""
}

// GetAvoidListStats returns statistics about the avoid list
func (s *Service) GetAvoidListStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Count by type
	tokenCount := 0
	walletCount := 0
	for _, entry := range s.entries {
		if entry.Type == "t" {
			tokenCount++
		} else if entry.Type == "w" {
			walletCount++
		}
	}

	return map[string]interface{}{
		"totalEntries": len(s.entries),
		"tokenCount":   tokenCount,
		"walletCount":  walletCount,
		"lastUpdated":  s.lastUpdated.Format(time.RFC3339),
	}
}
