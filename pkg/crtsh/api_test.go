// api_test.go
package crtsh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_SearchCertificates(t *testing.T) {
	// 创建测试服务器
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求参数
		if r.URL.Query().Get("searchtype") != "CN" || r.URL.Query().Get("common_name") != "example.com" {
			t.Errorf("Unexpected query parameters: %v", r.URL)
		}

		// 返回测试数据
		certs := []Certificate{{
			ID:           1,
			IssuerCAID:   123,
			RawNameValue: "example.com\n*.example.com",
			NotBefore:    time.Now().Add(-24 * time.Hour),
			NotAfter:     time.Now().Add(365 * 24 * time.Hour),
			SerialNumber: "ABCD1234",
		}}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(certs)
	}))
	defer testServer.Close()

	// 配置测试客户端
	client := NewClient()
	client.BaseURL = testServer.URL + "/"

	// 执行搜索
	params := QueryParams{
		SearchType: "CN",
		CN:         "example.com",
		PageSize:   10,
	}
	result, _, err := client.SearchCertificates(context.Background(), params)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// 验证结果
	if len(result) != 1 || result[0].ID != 1 {
		t.Errorf("Unexpected result: %+v", result)
	}
	if len(result[0].Domains) != 2 || result[0].Domains[0] != "example.com" {
		t.Errorf("Domain parsing failed: %v", result[0].Domains)
	}
}

func TestClient_HandleHTTPErrors(t *testing.T) {
	// 测试服务器返回500错误
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "server error"}`))
	}))
	defer testServer.Close()

	client := NewClient()
	client.BaseURL = testServer.URL + "/"

	_, _, err := client.SearchCertificates(context.Background(), QueryParams{})
	if err == nil || err.Error() != "api error (500): server error" {
		t.Errorf("Expected server error, got: %v", err)
	}
}

func TestClient_RetryLogic(t *testing.T) {
	retryCount := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if retryCount < 2 {
			retryCount++
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Certificate{})
	}))
	defer testServer.Close()

	client := NewClient()
	client.BaseURL = testServer.URL + "/"
	client.RetryCount = 3

	_, _, err := client.SearchCertificates(context.Background(), QueryParams{})
	if err != nil {
		t.Errorf("Retry failed: %v", err)
	}
	if retryCount != 2 {
		t.Errorf("Expected 2 retries, got %d", retryCount)
	}
}

func TestGetCertificateByID(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" || r.URL.Query().Get("id") != "123" {
			t.Errorf("Unexpected request: %v", r.URL)
		}
		cert := Certificate{ID: 123, SerialNumber: "TEST123"}
		json.NewEncoder(w).Encode(cert)
	}))
	defer testServer.Close()

	client := NewClient()
	client.BaseURL = testServer.URL + "/"

	cert, err := client.GetCertificateByID(context.Background(), 123)
	if err != nil {
		t.Fatal(err)
	}
	if cert.ID != 123 || cert.SerialNumber != "TEST123" {
		t.Errorf("Unexpected certificate data: %+v", cert)
	}
}
