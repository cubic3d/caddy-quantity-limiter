package quantitylimiter

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type NextHandlerOK struct{}

func (NextHandlerOK) ServeHTTP(w http.ResponseWriter, _ *http.Request) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func TestNonInteraction(t *testing.T) {
	ql := QuantityLimiter{Quantity: 2}
	if err := ql.Provision(caddy.Context{}); err != nil {
		t.Fatalf("could not provision module: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	if err := ql.ServeHTTP(rec, req, NextHandlerOK{}); err != nil {
		t.Fatalf("serving HTTP failed: %v", err)
	}

	if rec.Code != 200 {
		t.Fatalf("wrong response code: %d", rec.Code)
	}
}

func TestInitialDenial(t *testing.T) {
	ql := QuantityLimiter{Quantity: 2}
	if err := ql.Provision(caddy.Context{}); err != nil {
		t.Fatalf("could not provision module: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?"+ql.paramGet+"=test", nil)

	if err := ql.ServeHTTP(rec, req, NextHandlerOK{}); err != nil {
		t.Fatalf("serving HTTP failed: %v", err)
	}

	if rec.Code != 404 {
		t.Fatalf("wrong response code: %d", rec.Code)
	}
}

func TestGetSet(t *testing.T) {
	ql := QuantityLimiter{Quantity: 2}
	if err := ql.Provision(caddy.Context{}); err != nil {
		t.Fatalf("could not provision module: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?"+ql.paramSet+"=test", nil)

	if err := ql.ServeHTTP(rec, req, NextHandlerOK{}); err != nil {
		t.Fatalf("serving HTTP failed: %v", err)
	}

	if rec.Code != 202 {
		t.Fatalf("wrong response code: %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/test?"+ql.paramGet+"=test", nil)

	if err := ql.ServeHTTP(rec, req, NextHandlerOK{}); err != nil {
		t.Fatalf("serving HTTP failed: %v", err)
	}

	if rec.Code != 200 {
		t.Fatalf("wrong response code: %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/test?"+ql.paramGet+"=test", nil)

	if err := ql.ServeHTTP(rec, req, NextHandlerOK{}); err != nil {
		t.Fatalf("serving HTTP failed: %v", err)
	}

	if rec.Code != 200 {
		t.Fatalf("wrong response code: %d", rec.Code)
	}

	if req.Header.Get(ql.paramGet) != "" {
		t.Fatalf("found module header in response")
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/test?"+ql.paramGet+"=test", nil)

	if err := ql.ServeHTTP(rec, req, NextHandlerOK{}); err != nil {
		t.Fatalf("serving HTTP failed: %v", err)
	}

	if rec.Code != 404 {
		t.Fatalf("wrong response code: %d", rec.Code)
	}
}

func TestUnmarshalCandyfile(t *testing.T) {
	directive := `quantity_limiter {
  parameterNamePrefix prefix
  quantity 2
}`

	d := caddyfile.NewTestDispenser(directive)
	ql := QuantityLimiter{}
	if err := ql.UnmarshalCaddyfile(d); err != nil {
		t.Fatalf("failed parsing Candyfile %v", err)
	}

	expect := QuantityLimiter{
		ParameterNamePrefix: "prefix",
		Quantity:            2,
	}

	if !reflect.DeepEqual(ql, expect) {
		t.Fatal("unexpected configuration in module")
	}
}
