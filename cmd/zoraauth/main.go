package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

// Struct to parse the device code response
type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_URI"`
	VerificationURIComplete string `json:"verification_URI_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// Struct to parse the token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

type TokenData struct {
	Domain       string `yaml:"domain"`
	ClientID     string `yaml:"clientId"`
	AccessToken  string `yaml:"accessToken"`
	RefreshToken string `yaml:"refreshToken"`
	TokenType    string `yaml:"tokenType"`
}

type ZoraAuthResponse struct {
	ZoraAuth *TokenData `json:"zoraauth" yaml:"zoraauth"`
}

// Function to request the device code
func requestDeviceCode(domain, clientID, audience string) (*DeviceCodeResponse, error) {
	url := fmt.Sprintf("https://%s/oauth/device/code", domain)
	data := fmt.Sprintf("client_id=%s&scope=profile%%20email%%20offline_access%%20openid&audience=%s", clientID, audience)

	resp, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewBufferString(data))
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var deviceCodeResponse DeviceCodeResponse
	if err := json.Unmarshal(body, &deviceCodeResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &deviceCodeResponse, nil
}

// Function to poll for token
func pollForToken(domain, clientID, deviceCode string, interval, expiresIn int) (*TokenData, error) {
	url := fmt.Sprintf("https://%s/oauth/token", domain)
	data := fmt.Sprintf("client_id=%s&grant_type=urn:ietf:params:oauth:grant-type:device_code&device_code=%s", clientID, deviceCode)

	timer := 0
	for timer < expiresIn {
		time.Sleep(time.Duration(interval) * time.Second)
		timer += interval

		resp, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewBufferString(data))
		if err != nil {
			return nil, fmt.Errorf("failed to poll for token: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response body: %w", err)
			}

			var tokenResponse TokenResponse
			if err := json.Unmarshal(body, &tokenResponse); err != nil {
				return nil, fmt.Errorf("failed to parse JSON response: %w", err)
			}
			tokenData := TokenData{
				Domain:       domain,
				ClientID:     clientID,
				AccessToken:  tokenResponse.AccessToken,
				RefreshToken: tokenResponse.RefreshToken,
				TokenType:    tokenResponse.TokenType,
			}
			return &tokenData, nil
		}
	}

	return nil, errors.New("failed to retrieve tokens within the expiration time")
}

// Function to write token information to a YAML file
func writeTokensToYaml(filename string, tokens *TokenData) error {
	oauth := ZoraAuthResponse{tokens}
	yamlData, err := yaml.Marshal(oauth)
	if err != nil {
		return fmt.Errorf("failed to marshal tokens to YAML: %w", err)
	}

	if err := os.WriteFile(filename, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write tokens to file: %w", err)
	}

	return nil
}

func main() {
	// Define command-line flags using pflag
	domain := pflag.String("domain", "", "OAuth domain (e.g. Auth0 domain)")
	clientID := pflag.String("client-id", "", "OAuth client ID")
	audience := pflag.String("audience", "", "OAuth audience")
	outputFile := pflag.String("output", "tokens.yaml", "Output file for tokens in YAML format")

	// Parse flags
	pflag.Parse()

	// Ensure required flags are provided
	if *domain == "" || *clientID == "" || *audience == "" {
		fmt.Println("Error: domain, client-id, and audience must be provided.")
		pflag.Usage()
		os.Exit(1)
	}

	// Step 1: Request device code
	fmt.Println("Initiating Device Authorization Flow...")
	deviceInfo, err := requestDeviceCode(*domain, *clientID, *audience)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Display instructions to the user
	fmt.Printf("Please visit %s and enter code: %s, or visit: %s\n",
		deviceInfo.VerificationURI, deviceInfo.UserCode, deviceInfo.VerificationURIComplete)

	// Step 3: Poll for token
	tokens, err := pollForToken(*domain, *clientID, deviceInfo.DeviceCode, deviceInfo.Interval, deviceInfo.ExpiresIn)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(2)
	}

	// Step 4: Output tokens to a YAML file
	if err := writeTokensToYaml(*outputFile, tokens); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(3)
	}

	fmt.Printf("Tokens saved to %s\n", *outputFile)
}
