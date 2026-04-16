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

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"items":[],"total":0}`)),
			}, nil
		}),
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/articles?q=kubernetes", nil)

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
			if req.URL.Path != "/api/v1/sources/12" {
				t.Fatalf("unexpected path: %s", req.URL.Path)
			}
			if got := req.Header.Get("Content-Type"); got != "application/json" {
				t.Fatalf("unexpected content type: %s", got)
			}

			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			if string(body) != `{"is_enabled":false}` {
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
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/sources/12", strings.NewReader(`{"is_enabled":false}`))
	c.Request.Header.Set("Content-Type", "application/json")

	proxyRequest(c, client, "http://article-service/api/v1/sources/12")

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got=%d want=%d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != `{"id":12}` {
		t.Fatalf("unexpected body: %s", body)
	}
}
