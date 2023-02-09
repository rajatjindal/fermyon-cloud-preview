package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

type App struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	StorageID   string    `json:"storageId"`
	Description string    `json:"description"`
	Channels    []Channel `json:"channels"`
}

type Channel struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	ActiveRevisionNumber string    `json:"activeRevisionNumber"`
	Domain               string    `json:"domain"`
	Created              time.Time `json:"created"`
}
type GetAppsResponse struct {
	Apps       []App `json:"items"`
	TotalItems int   `json:"totalItems"`
	PageIndex  int   `json:"pageIndex"`
	PageSize   int   `json:"pageSize"`
	IsLastPage bool  `json:"isLastPage"`
}

const (
	ProductionCloudLink = "https://cloud.fermyon.com"
	TokenBaseDir        = "/home/runner/.config/fermyon"
)

type Client struct {
	Base       string
	Token      string
	httpclient *http.Client
}

type TokenInfo struct {
	Token string `json:"token"`
}

func getToken() (string, error) {
	tokenFile := filepath.Join(TokenBaseDir, os.Getenv("INPUT_FERMYON_DEPLOYMENT_ENV"))
	raw, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", err
	}

	var tokenInfo TokenInfo
	err = json.Unmarshal(raw, &tokenInfo)
	if err != nil {
		return "", err
	}

	return tokenInfo.Token, nil
}

func NewClient(base string) (*Client, error) {
	token, err := getToken()
	if err != nil {
		return nil, err
	}

	return &Client{
		Base:  base,
		Token: token,
		httpclient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

func (c *Client) GetAllApps() ([]App, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/apps", c.Base), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	rawresp, err := c.httpclient.Do(req)
	if err != nil {
		return nil, err
	}
	defer rawresp.Body.Close()

	rawbody, err := io.ReadAll(rawresp.Body)
	if err != nil {
		return nil, err
	}

	if rawresp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting apps for user. Expected status code: %d, got: %d. Body: %s", http.StatusOK, rawresp.StatusCode, string(rawbody))
	}

	var resp GetAppsResponse
	err = json.Unmarshal(rawbody, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Apps, nil
}

func (c *Client) getAppIdWithName(name string) (string, error) {
	apps, err := c.GetAllApps()
	if err != nil {
		return "", err
	}

	for _, app := range apps {
		if app.Name == name {
			return app.ID, nil
		}
	}

	return "", fmt.Errorf("no app found with name %s", name)
}

func (c *Client) deleteAppById(appId string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/apps/%s", c.Base, appId), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	resp, err := c.httpclient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rawbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("error activating user code. Expected status code: %d, got: %d. Body: %s", http.StatusNoContent, resp.StatusCode, string(rawbody))
	}

	return nil
}

func (c *Client) DeleteAppByName(appName string) error {
	appId, err := c.getAppIdWithName(appName)
	if err != nil {
		return err
	}

	return c.deleteAppById(appId)
}

func (c *Client) Deploy(appname string) (*Metadata, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("spin", "deploy")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		logrus.Errorf("cmd: %s\nstdout:\n %s stderr:\n %s", cmd.String(), stdout.String(), stderr.String())
		return nil, err
	}

	return ExtractMetadataFromLogs(appname, stdout.String())
}
