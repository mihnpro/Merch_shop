package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
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
	hasBody := body != nil
	if hasBody {
		buf, err := json.Marshal(body)
		if err != nil {
			return 0, nil, fmt.Errorf("marshal body: %w", err)
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequest(method, c.cfg.apiURL(path), reader)
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
	Login     string `json:"login"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type loginBody struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type createCategoryBody struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type createProductBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PricePoints int64  `json:"price_points"`
	CategoryID  string `json:"category_id"`
}

type updateProductBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PricePoints int64  `json:"price_points"`
	CategoryID  string `json:"category_id"`
	Active      bool   `json:"active"`
	Version     int    `json:"version"`
}

type adjustStockBody struct {
	ProductID   string `json:"product_id"`
	Delta       int    `json:"delta"`
	OperationID string `json:"operation_id"`
	Reason      string `json:"reason"`
}

type addCartItemBody struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type updateCartItemBody struct {
	Quantity int `json:"quantity"`
}

type userView struct {
	ID    string `json:"id"`
	Login string `json:"login"`
	Role  string `json:"role"`
}

type categoryView struct {
	ID string `json:"id"`
}

type productView struct {
	ID          string `json:"id"`
	PricePoints int64  `json:"price_points"`
	Version     int    `json:"version"`
}

type cartItemView struct {
	ID           string `json:"id"`
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
	Quantity     int    `json:"quantity"`
	PricePerUnit int    `json:"price_per_unit"`
	LineTotal    int    `json:"line_total"`
}

type cartView struct {
	Items     []cartItemView `json:"items"`
	Total     int            `json:"total"`
	ItemCount int            `json:"item_count"`
}
