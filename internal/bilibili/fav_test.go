package bilibili

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestListCreatedFoldersPaginates(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Query().Get("up_mid") != "42" {
			t.Errorf("up_mid = %s", r.URL.Query().Get("up_mid"))
		}
		if r.Header.Get("Cookie") == "" {
			t.Error("missing cookie")
		}
		switch r.URL.Query().Get("pn") {
		case "1":
			_, _ = w.Write([]byte(`{"code":0,"data":{"list":[{"id":11,"title":"第一页","cover":"c1","ctime":1499900654,"mtime":1523352284,"media_count":3}],"has_more":true}}`))
		case "2":
			_, _ = w.Write([]byte(`{"code":0,"data":{"list":[{"id":12,"title":"第二页","cover":"","ctime":1,"mtime":2,"media_count":0}],"has_more":false}}`))
		default:
			t.Errorf("unexpected page %s", r.URL.Query().Get("pn"))
		}
	}))
	defer server.Close()

	cookies := []Cookie{{Name: "DedeUserID", Value: "42"}}
	client := &Client{BaseURL: server.URL, CookieHeader: "DedeUserID=42", HTTP: server.Client()}
	folders, err := ListCreatedFolders(context.Background(), client, cookies, 0)
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 || len(folders) != 2 || folders[0].ID != 11 || folders[1].Title != "第二页" {
		t.Fatalf("hits=%d folders=%+v", hits.Load(), folders)
	}
}

func TestListCreatedFoldersRequiresUserID(t *testing.T) {
	_, err := ListCreatedFolders(context.Background(), nil, nil, 0)
	if err == nil {
		t.Fatal("expected missing DedeUserID")
	}
}

func TestListCreatedFoldersRejectsEmptyData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":null}`))
	}))
	defer server.Close()
	client := &Client{BaseURL: server.URL, HTTP: server.Client()}
	_, err := ListCreatedFolders(context.Background(), client, []Cookie{{Name: "DedeUserID", Value: "1"}}, 0)
	if err == nil {
		t.Fatal("expected empty data error")
	}
}

func TestListFavoriteMediasPaginates(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Query().Get("media_id") != "7" || r.URL.Query().Get("platform") != "web" {
			t.Errorf("query = %s", r.URL.RawQuery)
		}
		switch r.URL.Query().Get("pn") {
		case "1":
			_, _ = w.Write([]byte(`{"code":0,"data":{"medias":[{"id":11,"title":"一","bvid":"BV1","pubtime":1,"fav_time":2},null],"has_more":true}}`))
		case "2":
			_, _ = w.Write([]byte(`{"code":0,"data":{"medias":[{"id":12,"title":"二","bvid":"BV2","pubtime":3,"fav_time":4}],"has_more":false}}`))
		default:
			t.Errorf("page %s", r.URL.Query().Get("pn"))
		}
	}))
	defer server.Close()
	client := &Client{BaseURL: server.URL, HTTP: server.Client()}
	medias, err := ListFavoriteMedias(context.Background(), client, 7, 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 || len(medias) != 2 || medias[1].Bvid != "BV2" {
		t.Fatalf("hits=%d medias=%+v", hits.Load(), medias)
	}
}

func TestParseNotifiesCookieExpired(t *testing.T) {
	var got error
	client := &Client{OnError: func(err error) { got = err }}
	_, _, err := client.Parse([]byte(`{"code":-101,"message":"账号未登录","data":null}`))
	if !IsCookieExpired(err) || !IsCookieExpired(got) {
		t.Fatalf("err=%v got=%v", err, got)
	}
}

func TestAPICookieExpired(t *testing.T) {
	_, _, err := APIData([]byte(`{"code":-101,"message":"账号未登录","data":null}`))
	if !IsCookieExpired(err) {
		t.Fatal(err)
	}
}
