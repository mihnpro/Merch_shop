package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
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
	return c.send(req)
}

// uploadPhoto sends a multipart/form-data request to the media endpoint.
// fieldName lets tests exercise the "no file field" case; partContentType is
// the Content-Type header set on the part (what the service validates).
func (c *Client) uploadPhoto(fieldName, partContentType string, data []byte) (int, []byte, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename="photo.bin"`, fieldName))
	if partContentType != "" {
		h.Set("Content-Type", partContentType)
	}
	part, err := w.CreatePart(h)
	if err != nil {
		return 0, nil, err
	}
	if _, err := part.Write(data); err != nil {
		return 0, nil, err
	}
	if err := w.Close(); err != nil {
		return 0, nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.cfg.apiURL("/admin/media/photos"), &buf)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return c.send(req)
}

func (c *Client) send(req *http.Request) (int, []byte, error) {
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

type userView struct {
	ID    string `json:"id"`
	Login string `json:"login"`
	Role  string `json:"role"`
}

type uploadView struct {
	PhotoKey string `json:"photo_key"`
}
