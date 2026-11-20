package api

import (
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMustMarshalDoesNotPanicOnUnmarshalable(t *testing.T) {
	// math.Inf is not representable in JSON, so json.Marshal fails. The old
	// implementation panicked; MustMarshal must now return an error envelope.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MustMarshal panicked: %v", r)
		}
	}()
	out := MustMarshal(math.Inf(1))
	if !strings.Contains(out, "error") {
		t.Fatalf("expected an error envelope, got %q", out)
	}
}

func TestMustMarshalValidUnchanged(t *testing.T) {
	if got := MustMarshal(map[string]int{"a": 1}); got != `{"a":1}` {
		t.Fatalf("valid marshal changed: got %q", got)
	}
}

func TestHealthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterHealthRoutes(router)
	for _, path := range []string{"/healthz", "/readyz"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s: got status %d, want 200", path, w.Code)
		}
	}
}

func TestRespondError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	RespondError(c, http.StatusBadRequest, errors.New("boom"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"error":"boom"`) {
		t.Fatalf("unexpected body %q", w.Body.String())
	}
}

func TestParseServerConfig(t *testing.T) {
	def := parseServerConfig(nil)
	if def.addr != ":8080" || def.disableAuth || def.tlsCert != "" || def.tlsKey != "" {
		t.Fatalf("unexpected defaults: %+v", def)
	}
	got := parseServerConfig([]string{"--disable-auth", "--addr=127.0.0.1:9000", "--tls-cert=c.pem", "--tls-key=k.pem"})
	if got.addr != "127.0.0.1:9000" || !got.disableAuth || got.tlsCert != "c.pem" || got.tlsKey != "k.pem" {
		t.Fatalf("unexpected parse: %+v", got)
	}
	// Unknown/extra args must not crash startup.
	_ = parseServerConfig([]string{"--totally-unknown", "positional"})
}
