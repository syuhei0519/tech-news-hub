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

func TestProxyGetForwardsQueryAndBody(t *testing.T) {
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

	proxyGet(c, client, "http://article-service/api/v1/articles")

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

func TestProxyGetReturnsBadGatewayOnTransportError(t *testing.T) {
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

	proxyGet(c, client, "http://article-service/api/v1/articles")

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status code: got=%d want=%d", rec.Code, http.StatusBadGateway)
	}
}
