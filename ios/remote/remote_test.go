package remote

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseSize(t *testing.T) {
	cases := []struct {
		name         string
		in           string
		wantW, wantH float64
		wantErr      bool
	}{
		{"wda envelope", `{"value":{"width":390,"height":844}}`, 390, 844, false},
		{"bare object", `{"width":320,"height":568}`, 320, 568, false},
		{"leading noise", "some log line\n{\"value\":{\"width\":428,\"height\":926}}", 428, 926, false},
		{"non positive", `{"value":{"width":0,"height":0}}`, 0, 0, true},
		{"garbage", `not json`, 0, 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w, h, err := parseSize([]byte(c.in))
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v x %v", w, h)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if w != c.wantW || h != c.wantH {
				t.Fatalf("got %vx%v, want %vx%v", w, h, c.wantW, c.wantH)
			}
		})
	}
}

func TestFractionToPoints(t *testing.T) {
	// Prime the cache so fractionToPoints doesn't shell out to `ios ui size`.
	s := &Server{sizeW: 390, sizeH: 844}

	cases := []struct {
		fx, fy       float64
		wantX, wantY float64
	}{
		{0, 0, 0, 0},
		{1, 1, 390, 844},
		{0.5, 0.5, 195, 422},
		{-0.5, 2, 0, 844}, // clamped to [0,1]
	}
	for _, c := range cases {
		x, y, err := s.fractionToPoints(c.fx, c.fy)
		if err != nil {
			t.Fatalf("fractionToPoints(%v,%v): %v", c.fx, c.fy, err)
		}
		if x != c.wantX || y != c.wantY {
			t.Fatalf("fractionToPoints(%v,%v)=%v,%v want %v,%v", c.fx, c.fy, x, y, c.wantX, c.wantY)
		}
	}
}

func TestFtoa(t *testing.T) {
	// `ios ui tap` parses --x/--y as integers, so coordinates must be integer
	// strings (decimals are rejected).
	cases := map[float64]string{
		195:   "195",
		422.4: "422",
		422.5: "423", // rounds to nearest
		0:     "0",
	}
	for in, want := range cases {
		if got := ftoa(in); got != want {
			t.Fatalf("ftoa(%v)=%q, want %q", in, got, want)
		}
	}
}

func TestIndexServesHTML(t *testing.T) {
	s := &Server{} // handleIndex needs no device state
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	s.handleIndex(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("content-type = %q, want text/html", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<html") || !strings.Contains(body, `src="/screen"`) {
		t.Fatalf("body does not look like the remote UI: %.120q", body)
	}
}

func TestButtonRejectsUnknown(t *testing.T) {
	s := &Server{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/button", strings.NewReader(`{"name":"selfdestruct"}`))
	s.handleButton(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for unknown button", rec.Code)
	}
}
