package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	cfg   Config
	http  *http.Client
	token string
}

func newClient(cfg Config) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		cfg:  cfg,
		http: &http.Client{Jar: jar, Timeout: 15 * time.Second},
	}
}

func (c *Client) do(method, path string, body any) (int, []byte, error) {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return 0, nil, fmt.Errorf("marshal body: %w", err)
		}
		reader = bytes.NewReader(buf)
	}
	return c.send(method, path, reader, body != nil)
}

func (c *Client) doRaw(method, path, rawBody string) (int, []byte, error) {
	return c.send(method, path, strings.NewReader(rawBody), true)
}

func (c *Client) send(method, path string, body io.Reader, hasBody bool) (int, []byte, error) {
	req, err := http.NewRequest(method, c.cfg.apiURL(path), body)
	if err != nil {
		return 0, nil, fmt.Errorf("build request: %w", err)
	}
	if hasBody {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("read body: %w", err)
	}
	return resp.StatusCode, raw, nil
}

func (c *Client) login(login, password string) (int, []byte, error) {
	status, raw, err := c.do(http.MethodPost, "/auth/login", loginBody{Login: login, Password: password})
	if err != nil {
		return status, raw, err
	}
	if status == http.StatusOK {
		c.token = c.cookieValue(accessCookieName)
	}
	return status, raw, nil
}

func (c *Client) cookieValue(name string) string {
	u, err := url.Parse(c.cfg.GatewayURL)
	if err != nil {
		return ""
	}
	for _, ck := range c.http.Jar.Cookies(u) {
		if ck.Name == name {
			return ck.Value
		}
	}
	return ""
}

const accessCookieName = "access_token"

type registerBody struct {
	Login       string `json:"login"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Patronymic  string `json:"patronymic,omitempty"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

type loginBody struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type updateProfileBody struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Patronymic  string `json:"patronymic,omitempty"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

type changePasswordBody struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type grantPointsBody struct {
	Amount      int64  `json:"amount"`
	OperationID string `json:"operation_id"`
	Reason      string `json:"reason"`
}

type blockUserBody struct {
	Blocked bool   `json:"blocked"`
	Reason  string `json:"reason,omitempty"`
}

type changeRoleBody struct {
	Role string `json:"role"`
}

type userView struct {
	ID          string `json:"id"`
	Login       string `json:"login"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Patronymic  string `json:"patronymic"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
	Status      string `json:"status"`
}

type loginView struct {
	User userView `json:"user"`
}

type meView struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type balanceView struct {
	Points int64 `json:"points"`
}

type transactionsView struct {
	Transactions []struct {
		ID          string `json:"id"`
		OperationID string `json:"operation_id"`
		Amount      int64  `json:"amount"`
		Reason      string `json:"reason"`
	} `json:"transactions"`
	NextPageToken string `json:"next_page_token"`
}

type listUsersView struct {
	Users         []userView `json:"users"`
	NextPageToken string     `json:"next_page_token"`
}

type resetPasswordView struct {
	NewPassword string `json:"new_password"`
}

type statsView struct {
	NewUsers int `json:"new_users"`
}
