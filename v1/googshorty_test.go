package googshorty_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/orijtech/googshorty/v1"
)

func TestExpand(t *testing.T) {
	client, err := googshorty.NewClient(apiKey1)
	if err != nil {
		t.Fatalf("initializing the client: %v", err)
	}
	client.SetHTTPRoundTripper(&backend{route: expandRoute})

	tests := [...]struct {
		shortURL string
		want     *googshorty.URLDetails
		wantErr  bool
	}{
		0: {
			shortURL: "https://goo.gl/Zu6ATj",
			want:     urlDetailsFromFile("orijtech"),
		},

		1: {
			shortURL: "https://goo.gl/5ycdVx",
			want:     urlDetailsFromFile("medisa"),
		},

		// Test malformed URLs
		2: {
			shortURL: "",
			wantErr:  true,
		},
		3: {
			shortURL: "     ",
			wantErr:  true,
		},
		4: {
			// Unknown
			shortURL: fmt.Sprintf("%v", time.Now().Unix()),
			wantErr:  true,
		},
	}

	for i, tt := range tests {
		udetails, err := client.Expand(tt.shortURL)
		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d expecting non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d err: %v", i, err)
			continue
		}

		if udetails == nil {
			t.Errorf("#%d expecting url details back", i)
			continue
		}

		gotBlob, wantBlob := jsonMarshal(udetails), jsonMarshal(tt.want)
		if !bytes.Equal(gotBlob, wantBlob) {
			t.Errorf("#%d\ngot:  %s\nwant: %s", i, gotBlob, wantBlob)
		}
	}
}

func TestLookupAnalytics(t *testing.T) {
	client, err := googshorty.NewClient(apiKey1)
	if err != nil {
		t.Fatalf("initializing the client: %v", err)
	}
	client.SetHTTPRoundTripper(&backend{route: analyticsRoute})

	tests := [...]struct {
		url     string
		want    *googshorty.Analytics
		wantErr bool
	}{
		0: {
			url:  "https://goo.gl/XRdHKo",
			want: analyticsFromFile("googshorty"),
		},

		// Test malformed URLs
		1: {
			url:     "",
			wantErr: true,
		},
		2: {
			url:     "     ",
			wantErr: true,
		},
		3: {
			// Not a recognized URL
			url:     "/v2/flux",
			wantErr: true,
		},
	}

	for i, tt := range tests {
		analytics, err := client.LookupAnalytics(tt.url)
		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d expecting non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d err: %v", i, err)
			continue
		}

		if analytics == nil {
			t.Errorf("#%d expecting analytics back", i)
			continue
		}

		gotBlob, wantBlob := jsonMarshal(analytics), jsonMarshal(tt.want)
		if !bytes.Equal(gotBlob, wantBlob) {
			t.Errorf("#%d\ngot:  %s\nwant: %s", i, gotBlob, wantBlob)
		}
	}
}

func TestShorten(t *testing.T) {
	client, err := googshorty.NewClient(apiKey1)
	if err != nil {
		t.Fatalf("initializing the client: %v", err)
	}
	client.SetHTTPRoundTripper(&backend{route: shortenRoute})

	tests := [...]struct {
		longURL string
		want    *googshorty.URLDetails
		wantErr bool
	}{
		0: {
			longURL: "https://orijtech.com/",
			want:    urlDetailsFromFile("orijtech"),
		},

		1: {
			longURL: "https://medisa.orijtech.com/",
			want:    urlDetailsFromFile("medisa"),
		},

		// Test malformed URLs
		2: {
			longURL: "",
			wantErr: true,
		},
		3: {
			longURL: "     ",
			wantErr: true,
		},
		4: {
			// Not a recognized URL
			longURL: "/v2/flux",
			wantErr: true,
		},
	}

	for i, tt := range tests {
		udetails, err := client.Shorten(tt.longURL)
		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d expecting non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d err: %v", i, err)
			continue
		}

		if udetails == nil {
			t.Errorf("#%d expecting url details back", i)
			continue
		}

		gotBlob, wantBlob := jsonMarshal(udetails), jsonMarshal(tt.want)
		if !bytes.Equal(gotBlob, wantBlob) {
			t.Errorf("#%d\ngot:  %s\nwant: %s", i, gotBlob, wantBlob)
		}
	}
}

func urlDetailsPath(id string) string {
	return fmt.Sprintf("./testdata/url-details-%s.json", id)
}

func urlDetailsFromFile(id string) *googshorty.URLDetails {
	path := urlDetailsPath(id)
	save := new(googshorty.URLDetails)
	if err := jsonDeserializeFromPath(path, save); err != nil {
		return nil
	}
	return save
}

func analyticsPath(id string) string {
	return fmt.Sprintf("./testdata/analytics-%s.json", id)
}

func analyticsFromFile(id string) *googshorty.Analytics {
	path := analyticsPath(id)
	save := new(googshorty.Analytics)
	if err := jsonDeserializeFromPath(path, save); err != nil {
		return nil
	}
	return save
}

func jsonDeserializeFromPath(path string, save interface{}) error {
	blob, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(blob, save)
}

func jsonMarshal(v interface{}) []byte {
	blob, _ := json.Marshal(v)
	return blob
}

type backend struct {
	route string
}

var _ http.RoundTripper = (*backend)(nil)

func (b *backend) RoundTrip(req *http.Request) (*http.Response, error) {
	switch b.route {
	case shortenRoute:
		return b.shortenRoundTrip(req)
	case expandRoute:
		return b.expandRoundTrip(req)
	case analyticsRoute:
		return b.analyticsRoundTrip(req)
	default:
		msg := fmt.Sprintf("%s is an unknown route", b.route)
		return makeResp(msg, http.StatusNotFound, nil), nil
	}
}

var hostToShortURLMapping = map[string]string{
	"orijtech.com":        "orijtech",
	"medisa.orijtech.com": "medisa",

	"Zu6ATj": "orijtech",
	"5ycdVx": "medisa",

	"XRdHKo": "googshorty",
}

func (b *backend) expandRoundTrip(req *http.Request) (*http.Response, error) {
	if badAuthResp, err := b.checkAuthAndMethod(req, "GET"); err != nil || badAuthResp != nil {
		return badAuthResp, err
	}
	query := req.URL.Query()
	shortURL := strings.TrimSpace(query.Get("shortUrl"))
	if shortURL == "" {
		msg := `expecting "shortUrl" in the query string`
		return makeResp(msg, http.StatusBadRequest, nil), nil
	}

	// Note that Google itself as of
	// Tue  9 May 2017 22:23:45 MDT
	// allows any kind of value shortened
	// so being strict here is of no value.
	// e.g see:
	//    `https://goo.gl/f0RWxs` which I set to just `github`
	// When they become strict, then sure uncomment this code.
	// if err != nil {
	// 	return makeResp(err.Error(), http.StatusBadRequest, nil), nil
	// }
	key := shortURL
	parsedURL, err := url.Parse(shortURL)
	if err == nil && parsedURL != nil {
		key = strings.TrimLeft(parsedURL.Path, "/")
	}

	shortURLID := hostToShortURLMapping[key]
	diskPath := urlDetailsPath(shortURLID)
	return makeResponseFromFile(diskPath)
}

func (b *backend) analyticsRoundTrip(req *http.Request) (*http.Response, error) {
	if badAuthResp, err := b.checkAuthAndMethod(req, "GET"); err != nil || badAuthResp != nil {
		return badAuthResp, err
	}
	query := req.URL.Query()
	shortURL := strings.TrimSpace(query.Get("shortUrl"))
	if shortURL == "" {
		msg := `expecting "shortUrl" in the query string`
		return makeResp(msg, http.StatusBadRequest, nil), nil
	}

	key := shortURL
	parsedURL, err := url.Parse(shortURL)
	if err == nil && parsedURL != nil {
		key = strings.TrimLeft(parsedURL.Path, "/")
	}

	shortURLID := hostToShortURLMapping[key]
	diskPath := analyticsPath(shortURLID)
	return makeResponseFromFile(diskPath)

}

func (b *backend) shortenRoundTrip(req *http.Request) (*http.Response, error) {
	if badAuthResp, err := b.checkAuthAndMethod(req, "POST"); err != nil || badAuthResp != nil {
		return badAuthResp, err
	}
	slurp, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return makeResp(err.Error(), http.StatusBadRequest, nil), nil
	}
	recv := make(map[string]string)
	if err := json.Unmarshal(slurp, &recv); err != nil {
		return makeResp(err.Error(), http.StatusBadRequest, nil), nil
	}
	longURL := recv["longUrl"]

	// Note that Google itself as of
	// Tue  9 May 2017 22:23:45 MDT
	// allows any kind of value shortened
	// so being strict here is of no value.
	// e.g see:
	//    `https://goo.gl/f0RWxs` which I set to just `github`
	// When they become strict, then sure uncomment this code.
	// if err != nil {
	// 	return makeResp(err.Error(), http.StatusBadRequest, nil), nil
	// }
	parsedURL, err := url.Parse(longURL)
	key := longURL
	if err == nil && parsedURL != nil {
		key = parsedURL.Host
	}

	shortURLID := hostToShortURLMapping[key]
	diskPath := urlDetailsPath(shortURLID)
	return makeResponseFromFile(diskPath)
}

func makeResponseFromFile(path string) (*http.Response, error) {
	f, err := os.Open(path)
	if err != nil {
		return makeResp(err.Error(), http.StatusInternalServerError, nil), nil
	}
	return makeResp("200 OK", http.StatusOK, f), nil
}

var knownAPIKeys = map[string]bool{
	apiKey1: true,
}

func authorizedAPIKey(apiKey string) bool {
	_, authorized := knownAPIKeys[apiKey]
	return authorized
}

func (b *backend) checkAuthAndMethod(req *http.Request, wantMethod string) (*http.Response, error) {
	if got, want := req.Method, wantMethod; got != want {
		msg := fmt.Sprintf("got method %q, want %q", got, want)
		return makeResp(msg, http.StatusMethodNotAllowed, nil), nil
	}

	query := req.URL.Query()
	apiKey := strings.TrimSpace(query.Get("key"))
	if !authorizedAPIKey(apiKey) {
		return makeResp("unauthorized api key", http.StatusUnauthorized, nil), nil
	}
	return nil, nil
}

func makeResp(status string, code int, body io.ReadCloser) *http.Response {
	return &http.Response{
		Status:     status,
		StatusCode: code,
		Body:       body,
		Header:     make(http.Header),
	}
}

const (
	shortenRoute   = "shorten"
	expandRoute    = "expand"
	analyticsRoute = "analytics"

	apiKey1 = "api-key-1"
)
