package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"HyLauncher/internal/game"
	"HyLauncher/pkg/model"
)

type AuthService struct {
	ctx     context.Context
	baseUrl string
}

func NewAuthService(ctx context.Context) *AuthService {
	return &AuthService{
		ctx:     ctx,
		baseUrl: "https://sessions.sanasol.ws",
	}
}

type gameSessionRequest struct {
	UUID   string   `json:"uuid"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

type gameSessionResponse struct {
	IdentityToken string    `json:"identityToken"`
	SessionToken  string    `json:"sessionToken"`
	ExpiresIn     int       `json:"expiresIn"`
	ExpiresAt     time.Time `json:"expiresAt"`
	TokenType     string    `json:"tokenType"`
}

func (s *AuthService) FetchGameSession(username string) (*model.GameSession, error) {
	url := s.baseUrl + "/game-session/new"

	uuid := game.OfflineUUID(username).String()

	reqBody := gameSessionRequest{
		UUID:   uuid,
		Name:   username,
		Scopes: []string{"hytale:server", "hytale:client"},
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	var gsResp gameSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&gsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &model.GameSession{
		Username:      username,
		UUID:          uuid,
		IdentityToken: gsResp.IdentityToken,
		SessionToken:  gsResp.SessionToken,
	}, nil
}
