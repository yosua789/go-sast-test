package helper

import (
	"assist-tix/dto"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func VerifyUserGarudaIDByID(baseURL, id, privateKey string) (*dto.ApiResponseGarudaIDService, error) {
	url := fmt.Sprintf("%s/v1/user/verify/%s", baseURL, id)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil) // Replace with actual URL
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	headerKey, err := HashBcryptKey(privateKey)
	if err != nil {
		log.Printf("Error hashing private key: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("X-API-Key", headerKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp *dto.ApiResponseGarudaIDService
	err = json.Unmarshal(bodyBytes, &apiResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return apiResp, nil
}
