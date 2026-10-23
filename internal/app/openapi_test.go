package app

import (
	"encoding/json"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"

	"huydq2801.com/golearn/internal/apidocs"
	"huydq2801.com/golearn/internal/config"
)

func TestOpenAPICoversEveryRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	application, err := NewApplication(&config.Config{RatePerSecond: 5, RateBurst: 10}, nil, discardLogger())
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err = json.Unmarshal(apidocs.Spec, &spec); err != nil {
		t.Fatal(err)
	}
	parameter := regexp.MustCompile(`:[^/]+`)
	for _, route := range application.routes.Routes() {
		if route.Path == "/openapi.json" {
			continue
		}
		path := parameter.ReplaceAllString(route.Path, "{uuid}")
		method := map[string]string{"GET": "get", "POST": "post", "PUT": "put", "PATCH": "patch", "DELETE": "delete"}[route.Method]
		if _, ok := spec.Paths[path][method]; !ok {
			t.Errorf("missing OpenAPI operation: %s %s", route.Method, path)
		}
	}
	for _, path := range []string{"/health/live", "/openapi.json"} {
		w := httptest.NewRecorder()
		application.Handler().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 {
			t.Errorf("%s status=%d", path, w.Code)
		}
	}
	w := httptest.NewRecorder()
	application.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/unknown", nil))
	if w.Code != 404 {
		t.Fatal(w.Code)
	}
}
