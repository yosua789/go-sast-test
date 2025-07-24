package helper

import (
	"assist-tix/dto"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func VerifyUserGarudaIDByID(baseURL, id string) (*dto.ApiResponseGarudaIDService, error) {
	url := fmt.Sprintf("%s/v1/user/verify/%s", baseURL, id)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make GET request: %w", err)
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
