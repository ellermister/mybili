package bilibili

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrEmptyCookie = errors.New("cookie content cannot be empty")

type Cookie struct {
	Domain string
	Path   string
	Name   string
	Value  string
}

func ParseNetscape(content string) ([]Cookie, error) {
	if strings.TrimSpace(content) == "" {
		return nil, ErrEmptyCookie
	}
	var cookies []Cookie
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 7 {
			continue
		}
		cookies = append(cookies, Cookie{
			Domain: parts[0],
			Path:   parts[2],
			Name:   parts[5],
			Value:  parts[6],
		})
	}
	return cookies, nil
}

func Header(cookies []Cookie) string {
	pairs := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie.Name == "" {
			continue
		}
		pairs = append(pairs, cookie.Name+"="+cookie.Value)
	}
	return strings.Join(pairs, "; ")
}

func LoggedIn(content string) (bool, error) {
	cookies, err := ParseNetscape(content)
	if err != nil {
		return false, err
	}
	req, err := http.NewRequest(http.MethodGet, "https://api.bilibili.com/x/web-interface/nav", nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://space.bilibili.com/")
	req.Header.Set("Cookie", Header(cookies))
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return false, err
	}
	var payload struct {
		Data struct {
			IsLogin bool `json:"isLogin"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return false, err
	}
	return payload.Data.IsLogin, nil
}
