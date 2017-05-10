package googshorty

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/orijtech/otils"
)

type URLDetails struct {
	Kind          string `json:"kind"`
	ShortURL      string `json:"id"`
	LongURL       string `json:"longUrl,omitempty"`
	StatusMessage string `json:"status"`
}

type CountIDPair struct {
	Count uint64 `json:"count,string"`
	ID    string `json:"id"`
}

const (
	baseURL = "https://www.googleapis.com/urlshortener/v1"
)

var (
	errBlankContent = errors.New("expecting non-blank content")
)

func (c *Client) Shorten(longURL string) (*URLDetails, error) {
	// Note that Google itself as of
	// Tue  9 May 2017 22:23:45 MDT
	// allows any kind of value shortened
	// so being strict here is of no value.
	// e.g see:
	//    `https://goo.gl/f0RWxs` which I set to just `github`
	// When they become strict, then sure uncomment this code.
	// parsedURL, err := url.Parse(longURL)
	// _, err := url.Parse(longURL)
	// if err != nil {
	// 	return nil, err
	// }
	// The closest to validation we can do here is
	// to reject purely whitespace and blank URLs.
	if strings.TrimSpace(longURL) == "" {
		return nil, errBlankContent
	}

	blob, err := json.Marshal(&URLDetails{LongURL: longURL})
	if err != nil {
		return nil, err
	}
	fullURL := fmt.Sprintf("%s/url?key=%s", baseURL, c.apiKey())
	req, err := http.NewRequest("POST", fullURL, bytes.NewReader(blob))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	blob, _, err = c.doAuthAndReq(req)
	if err != nil {
		return nil, err
	}
	udetails := new(URLDetails)
	if err := json.Unmarshal(blob, udetails); err != nil {
		return nil, err
	}
	return udetails, nil
}

func (c *Client) doAuthAndReq(req *http.Request) ([]byte, http.Header, error) {
	res, err := c.httpClient().Do(req)
	if err != nil {
		return nil, nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if !otils.StatusOK(res.StatusCode) {
		msg := res.Status
		if res.Body != nil {
			slurp, _ := ioutil.ReadAll(res.Body)
			if len(slurp) > 3 {
				msg = string(slurp)
			}
		}
		return nil, res.Header, errors.New(msg)
	}

	slurp, err := ioutil.ReadAll(res.Body)
	return slurp, res.Header, err
}

func (c *Client) Expand(shortURL string) (*URLDetails, error) {
	fullURL := fmt.Sprintf("%s/url?shortUrl=%s&key=%s", baseURL, shortURL, c.apiKey())
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	slurp, _, err := c.doAuthAndReq(req)
	if err != nil {
		return nil, err
	}

	uinf := new(URLDetails)
	if err := json.Unmarshal(slurp, uinf); err != nil {
		return nil, err
	}
	return uinf, nil
}

type Analytics struct {
	Kind      string     `json:"kind"`
	ID        string     `json:"id"`
	Status    string     `json:"status"`
	CreatedAt *time.Time `json:"created"`

	Analytics *Analytic `json:"analytics"`
}

type Analytic struct {
	AllTime          *AnalyticDetails `json:"allTime"`
	WithinLastMonth  *AnalyticDetails `json:"month"`
	WithinLastWeek   *AnalyticDetails `json:"week"`
	WithinLastDay    *AnalyticDetails `json:"day"`
	WithinLast2Hours *AnalyticDetails `json:"twoHours"`
}

type AnalyticDetails struct {
	ShortURLClicks uint64         `json:"shortUrlClicks,string"`
	LongURLClicks  uint64         `json:"longUrlClicks,string"`
	Referrers      []*CountIDPair `json:"referrers"`
	Countries      []*CountIDPair `json:"countries"`
	Browsers       []*CountIDPair `json:"browsers"`
	Platforms      []*CountIDPair `json:"platforms"`
}

var errUnimplemented = errors.New("unimplemented")

func (c *Client) LookupAnalytics(shortURL string) (*Analytics, error) {
	fullURL := fmt.Sprintf("%s/url?shortUrl=%s&key=%s&projection=FULL", baseURL, shortURL, c.apiKey())
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	slurp, _, err := c.doAuthAndReq(req)
	if err != nil {
		return nil, err
	}

	analytics := new(Analytics)
	if err := json.Unmarshal(slurp, analytics); err != nil {
		return nil, err
	}
	return analytics, nil
}
