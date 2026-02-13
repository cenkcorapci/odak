package mlflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client is a simple MLflow client
type Client struct {
	baseURL string
	client  *http.Client
}

// NewClient creates a new MLflow client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// ModelVersion represents a model version in MLflow
type ModelVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Source  string `json:"source"`
	RunID   string `json:"run_id"`
	Status  string `json:"status"`
}

// RegisterModel registers a model in MLflow
func (c *Client) RegisterModel(name, runID, modelPath string) error {
	url := fmt.Sprintf("%s/api/2.0/mlflow/registered-models/create", c.baseURL)
	
	payload := map[string]string{
		"name": name,
	}
	
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to register model: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}
	
	return nil
}

// GetModelVersion retrieves a model version
func (c *Client) GetModelVersion(name, version string) (*ModelVersion, error) {
	url := fmt.Sprintf("%s/api/2.0/mlflow/model-versions/get?name=%s&version=%s", 
		c.baseURL, name, version)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get model version: status=%d", resp.StatusCode)
	}
	
	var result struct {
		ModelVersion ModelVersion `json:"model_version"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &result.ModelVersion, nil
}

// LogMetric logs a metric to MLflow
func (c *Client) LogMetric(runID, key string, value float64, timestamp int64) error {
	url := fmt.Sprintf("%s/api/2.0/mlflow/runs/log-metric", c.baseURL)
	
	payload := map[string]interface{}{
		"run_id":    runID,
		"key":       key,
		"value":     value,
		"timestamp": timestamp,
	}
	
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to log metric: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}
	
	return nil
}
