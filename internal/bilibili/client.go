package bilibili

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	APIHost   = "https://api.bilibili.com"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"
)

// HTTPError 是 Bilibili 返回的非成功 HTTP 状态。412 和 429 视为风控。
type HTTPError struct {
	Status  int
	Snippet string
}

func (e *HTTPError) Error() string {
	if e.Snippet == "" {
		return fmt.Sprintf("bilibili http %d", e.Status)
	}
	return fmt.Sprintf("bilibili http %d: %s", e.Status, e.Snippet)
}

func IsRiskControl(err error) bool {
	httpErr, ok := err.(*HTTPError)
	if ok && (httpErr.Status == http.StatusPreconditionFailed || httpErr.Status == http.StatusTooManyRequests) {
		return true
	}
	apiErr, ok := err.(*APIError)
	return ok && (apiErr.Code == -412 || apiErr.Code == 412)
}

func IsVideoUnavailable(err error) bool {
	apiErr, ok := err.(*APIError)
	if !ok {
		return false
	}
	switch apiErr.Code {
	case -404, -403, 62002, 62004, 62012:
		return true
	default:
		return false
	}
}

// APIError 是 HTTP 成功但业务 code 不是 0。
type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("bilibili api code %d", e.Code)
	}
	return fmt.Sprintf("bilibili api code %d: %s", e.Code, e.Message)
}

func IsCookieExpired(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.Code == -101
}

// Client 只负责带 Cookie 的 GET。BaseURL 为空时用正式接口地址。
type Client struct {
	BaseURL      string
	CookieHeader string
	HTTP         *http.Client
	// OnError 在接口返回业务错误时调用，用来发 Cookie 过期通知。为空则不调用。
	OnError func(error)
}

func NewClient(cookies []Cookie) *Client {
	return &Client{
		BaseURL:      APIHost,
		CookieHeader: Header(cookies),
		HTTP:         &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = APIHost
	}
	target := base + path
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		target = path
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", "https://www.bilibili.com/")
	if c.CookieHeader != "" {
		req.Header.Set("Cookie", c.CookieHeader)
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &HTTPError{Status: resp.StatusCode, Snippet: snippet(body)}
	}
	return body, nil
}

// APIData 取出 code=0 时的 data。data 为 null 时 ok 为 false。
func APIData(body []byte) (json.RawMessage, bool, error) {
	var envelope struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, false, fmt.Errorf("bilibili response is not json: %w", err)
	}
	if envelope.Code != 0 {
		return nil, false, &APIError{Code: envelope.Code, Message: envelope.Message}
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return nil, false, nil
	}
	return envelope.Data, true, nil
}

// Parse 解析接口正文，并在有业务错误时调用 OnError。
func (c *Client) Parse(body []byte) (json.RawMessage, bool, error) {
	raw, ok, err := APIData(body)
	if c != nil && c.OnError != nil && err != nil {
		c.OnError(err)
	}
	return raw, ok, err
}

func snippet(body []byte) string {
	text := strings.TrimSpace(string(body))
	if len(text) > 180 {
		return text[:180]
	}
	return text
}
