package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestProxyRequestForwardsQueryAndBody(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v1/articles" {
				t.Fatalf("unexpected path: %s", req.URL.Path)
			}
			if got := req.URL.Query().Get("q"); got != "kubernetes" {
				t.Fatalf("unexpected query: %s", got)
			}
			if got := req.URL.Query().Get("is_read"); got != "false" {
				t.Fatalf("unexpected is_read: %s", got)
			}
			if got := req.URL.Query().Get("is_favorite"); got != "true" {
				t.Fatalf("unexpected is_favorite: %s", got)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"items":[],"total":0}`)),
			}, nil
		}),
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/articles?q=kubernetes&is_read=false&is_favorite=true", nil)

	proxyRequest(c, client, "http://article-service/api/v1/articles")

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d want=%d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != `{"items":[],"total":0}` {
		t.Fatalf("unexpected body: %s", body)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content type: %s", got)
	}
}

func TestProxyRequestReturnsBadGatewayOnTransportError(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, io.EOF
		}),
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/articles", nil)

	proxyRequest(c, client, "http://article-service/api/v1/articles")

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status code: got=%d want=%d", rec.Code, http.StatusBadGateway)
	}
}

func TestProxyRequestForwardsMethodHeadersAndBody(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodPatch {
				t.Fatalf("unexpected method: %s", req.Method)
			}
			if req.URL.Path != "/api/v1/articles/12/favorite-status" {
				t.Fatalf("unexpected path: %s", req.URL.Path)
			}
			if got := req.Header.Get("Content-Type"); got != "application/json" {
				t.Fatalf("unexpected content type: %s", got)
			}

			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			if string(body) != `{"is_favorite":true}` {
				t.Fatalf("unexpected body: %s", string(body))
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"id":12}`)),
			}, nil
		}),
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/articles/12/favorite-status", strings.NewReader(`{"is_favorite":true}`))
	c.Request.Header.Set("Content-Type", "application/json")

	proxyRequest(c, client, "http://article-service/api/v1/articles/12/favorite-status")

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d want=%d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != `{"id":12}` {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestProxyRequestForwardsFetchJobsQuery(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v1/fetch-jobs" {
				t.Fatalf("unexpected path: %s", req.URL.Path)
			}
			if got := req.URL.Query().Get("source_id"); got != "7" {
				t.Fatalf("unexpected source_id: %s", got)
			}
			if got := req.URL.Query().Get("status"); got != "failed" {
				t.Fatalf("unexpected status: %s", got)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"items":[],"total":0,"page":1,"page_size":20,"total_pages":0}`)),
			}, nil
		}),
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/fetch-jobs?source_id=7&status=failed", nil)

	proxyRequest(c, client, "http://article-service/api/v1/fetch-jobs")

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d want=%d", rec.Code, http.StatusOK)
	}
}

func TestProxyCSVDownloadSetsAttachmentHeaders(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/api/v1/articles/export.csv" {
				t.Fatalf("unexpected path: %s", req.URL.Path)
			}
			if got := req.URL.Query().Get("source_id"); got != "4" {
				t.Fatalf("unexpected source_id: %s", got)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/csv; charset=utf-8"}},
				Body:       io.NopCloser(strings.NewReader("title\nexample\n")),
			}, nil
		}),
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/exports/articles.csv?source_id=4", nil)

	proxyCSVDownload(c, client, "http://article-service/api/v1/articles/export.csv", "articles-20260416-000000.csv")

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d want=%d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/csv; charset=utf-8" {
		t.Fatalf("unexpected content type: %s", got)
	}
	if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="articles-20260416-000000.csv"` {
		t.Fatalf("unexpected content disposition: %s", got)
	}
}

func TestProxyCSVDownloadPassesThroughErrors(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"error":"invalid from"}`)),
			}, nil
		}),
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/exports/articles.csv?from=bad", nil)

	proxyCSVDownload(c, client, "http://article-service/api/v1/articles/export.csv", "articles.csv")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: got=%d want=%d", rec.Code, http.StatusBadRequest)
	}
	if body := rec.Body.String(); body != `{"error":"invalid from"}` {
		t.Fatalf("unexpected body: %s", body)
	}
}
