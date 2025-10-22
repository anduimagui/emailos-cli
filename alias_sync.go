package mailos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type JMAPSession struct {
	Username        string              `json:"username"`
	ApiURL          string              `json:"apiUrl"`
	DownloadURL     string              `json:"downloadUrl"`
	UploadURL       string              `json:"uploadUrl"`
	EventSourceURL  string              `json:"eventSourceUrl"`
	State           string              `json:"state"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	Accounts        map[string]JMAPAccount `json:"accounts"`
	PrimaryAccounts map[string]string   `json:"primaryAccounts"`
}

type JMAPAccount struct {
	Name                 string `json:"name"`
	IsPersonal          bool   `json:"isPersonal"`
	IsReadOnly          bool   `json:"isReadOnly"`
	AccountCapabilities map[string]interface{} `json:"accountCapabilities"`
}

type JMAPRequest struct {
	Using       []string      `json:"using"`
	MethodCalls []interface{} `json:"methodCalls"`
}

type JMAPResponse struct {
	MethodResponses []interface{} `json:"methodResponses"`
	SessionState    string        `json:"sessionState"`
}

type Identity struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	ReplyTo     string `json:"replyTo,omitempty"`
	BCC         string `json:"bcc,omitempty"`
	TextSignature string `json:"textSignature,omitempty"`
	HTMLSignature string `json:"htmlSignature,omitempty"`
	MayDelete   bool   `json:"mayDelete"`
}

type IdentityGetResponse struct {
	AccountID string     `json:"accountId"`
	State     string     `json:"state"`
	List      []Identity `json:"list"`
	NotFound  []string   `json:"notFound"`
}

func GetJMAPSession(token string) (*JMAPSession, error) {
	req, err := http.NewRequest("GET", "https://api.fastmail.com/jmap/session", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var session JMAPSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode session response: %v", err)
	}

	return &session, nil
}

func GetIdentities(token, apiURL, accountID string) ([]Identity, error) {
	requestBody := JMAPRequest{
		Using: []string{
			"urn:ietf:params:jmap:core",
			"urn:ietf:params:jmap:submission",
		},
		MethodCalls: []interface{}{
			[]interface{}{
				"Identity/get",
				map[string]interface{}{
					"accountId": accountID,
				},
				"0",
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var jmapResp JMAPResponse
	if err := json.NewDecoder(resp.Body).Decode(&jmapResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(jmapResp.MethodResponses) == 0 {
		return nil, fmt.Errorf("no method responses received")
	}

	methodResponse := jmapResp.MethodResponses[0].([]interface{})
	if len(methodResponse) < 2 {
		return nil, fmt.Errorf("invalid method response structure")
	}

	responseData, err := json.Marshal(methodResponse[1])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response data: %v", err)
	}

	var identityResponse IdentityGetResponse
	if err := json.Unmarshal(responseData, &identityResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal identity response: %v", err)
	}

	return identityResponse.List, nil
}

func SyncFastMailAliases(token string) error {
	session, err := GetJMAPSession(token)
	if err != nil {
		return fmt.Errorf("failed to get JMAP session: %v", err)
	}

	var primaryAccountID string
	for id, account := range session.Accounts {
		if account.IsPersonal {
			primaryAccountID = id
			break
		}
	}

	if primaryAccountID == "" {
		return fmt.Errorf("no personal account found")
	}

	identities, err := GetIdentities(token, session.ApiURL, primaryAccountID)
	if err != nil {
		return fmt.Errorf("failed to get identities: %v", err)
	}

	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load current config: %v", err)
	}

	fmt.Printf("Found %d identities from FastMail:\n", len(identities))
	for i, identity := range identities {
		fmt.Printf("  %d. %s <%s>\n", i+1, identity.Name, identity.Email)
	}

	var newAccounts []AccountConfig
	existingEmails := make(map[string]bool)

	existingEmails[config.Email] = true
	for _, acc := range config.Accounts {
		existingEmails[acc.Email] = true
	}

	for _, identity := range identities {
		if strings.ToLower(identity.Email) == strings.ToLower(config.Email) {
			continue
		}

		if !existingEmails[identity.Email] {
			newAccount := AccountConfig{
				Email:        identity.Email,
				Provider:     config.Provider,
				Password:     config.Password,
				FromName:     identity.Name,
				FromEmail:    identity.Email,
				Label:        "Synced from FastMail",
			}
			newAccounts = append(newAccounts, newAccount)
			fmt.Printf("  → Adding new alias: %s <%s>\n", identity.Name, identity.Email)
		} else {
			fmt.Printf("  → Already configured: %s\n", identity.Email)
		}
	}

	if len(newAccounts) == 0 {
		fmt.Println("✓ All FastMail identities are already configured")
		return nil
	}

	config.Accounts = append(config.Accounts, newAccounts...)

	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save updated config: %v", err)
	}

	fmt.Printf("✓ Successfully synced %d new aliases from FastMail\n", len(newAccounts))
	return nil
}

func TestFastMailJMAPConnection(token string) error {
	fmt.Println("Testing FastMail JMAP API connection...")

	session, err := GetJMAPSession(token)
	if err != nil {
		return fmt.Errorf("failed to connect to JMAP session: %v", err)
	}

	fmt.Printf("✓ Connected to FastMail JMAP API\n")
	fmt.Printf("  Username: %s\n", session.Username)
	fmt.Printf("  API URL: %s\n", session.ApiURL)

	var primaryAccountID string
	for id, account := range session.Accounts {
		fmt.Printf("  Account: %s (Personal: %t, ReadOnly: %t)\n", 
			account.Name, account.IsPersonal, account.IsReadOnly)
		if account.IsPersonal {
			primaryAccountID = id
		}
	}

	if primaryAccountID == "" {
		return fmt.Errorf("no personal account found")
	}

	identities, err := GetIdentities(token, session.ApiURL, primaryAccountID)
	if err != nil {
		return fmt.Errorf("failed to retrieve identities: %v", err)
	}

	fmt.Printf("✓ Found %d email identities:\n", len(identities))
	for i, identity := range identities {
		fmt.Printf("  %d. %s <%s>\n", i+1, identity.Name, identity.Email)
	}

	return nil
}