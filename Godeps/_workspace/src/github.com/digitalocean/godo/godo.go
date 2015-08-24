package godo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/google/go-querystring/query"
	headerLink "github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/tent/http-link-go"
)

const (
	libraryVersion = "0.1.0"
	defaultBaseURL = "https://api.digitalocean.com/"
	userAgent      = "godo/" + libraryVersion
	mediaType      = "application/json"

	headerRateLimit     = "X-RateLimit-Limit"
	headerRateRemaining = "X-RateLimit-Remaining"
	headerRateReset     = "X-RateLimit-Reset"
)

// Client manages communication with DigitalOcean V2 API.
type Client struct {
	// HTTP client used to communicate with the DO API.
	client *http.Client

	// Base URL for API requests.
	BaseURL *url.URL

	// User agent for client
	UserAgent string

	// Rate contains the current rate limit for the client as determined by the most recent
	// API call.
	Rate Rate

	// Services used for communicating with the API
	Actions        ActionsService
	Domains        DomainsService
	Droplets       DropletsService
	DropletActions DropletActionsService
	Images         ImagesService
	ImageActions   ImageActionsService
	Keys           KeysService
	Regions        RegionsService
	Sizes          SizesService
}

// ListOptions specifies the optional parameters to various List methods that
// support pagination.
type ListOptions struct {
	// For paginated result sets, page of results to retrieve.
	Page int `url:"page,omitempty"`

	// For paginated result sets, the number of results to include per page.
	PerPage int `url:"per_page,omitempty"`
}

// Response is a Digital Ocean response. This wraps the standard http.Response returned from DigitalOcean.
type Response struct {
	*http.Response

	// Links that were returned with the response. These are parsed from
	// request body and not the header.
	Links *Links

	// Monitoring URI
	Monitor string

	Rate
}

// An ErrorResponse reports the error caused by an API request
type ErrorResponse struct {
	// HTTP response that caused this error
	Response *http.Response

	// Error message
	Message string
}

// Rate contains the rate limit for the current client.
type Rate struct {
	// The number of request per hour the client is currently limited to.
	Limit int `json:"limit"`

	// The number of remaining requests the client can make this hour.
	Remaining int `json:"remaining"`

	// The time at w\hic the current rate limit will reset.
	Reset Timestamp `json:"reset"`
}

func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)

	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qv, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	u.RawQuery = qv.Encode()
	return u.String(), nil
}

// NewClient returns a new Digital Ocean API client.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{client: httpClient, BaseURL: baseURL, UserAgent: userAgent}
	c.Actions = &ActionsServiceOp{client: c}
	c.Domains = &DomainsServiceOp{client: c}
	c.Droplets = &DropletsServiceOp{client: c}
	c.DropletActions = &DropletActionsServiceOp{client: c}
	c.Images = &ImagesServiceOp{client: c}
	c.ImageActions = &ImageActionsServiceOp{client: c}
	c.Keys = &KeysServiceOp{client: c}
	c.Regions = &RegionsServiceOp{client: c}
	c.Sizes = &SizesServiceOp{client: c}

	return c
}

// NewRequest creates an API request. A relative URL can be provided in urlStr, which will be resolved to the
// BaseURL of the Client. Relative URLS should always be specified without a preceding slash. If specified, the
// value pointed to by body is JSON encoded and included in as the request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	buf := new(bytes.Buffer)
	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", mediaType)
	req.Header.Add("Accept", mediaType)
	req.Header.Add("User-Agent", userAgent)
	return req, nil
}

// newResponse creates a new Response for the provided http.Response
func newResponse(r *http.Response) *Response {
	response := Response{Response: r}
	response.populateRate()

	return &response
}

func (r *Response) links() (map[string]headerLink.Link, error) {
	if linkText, ok := r.Response.Header["Link"]; ok {
		links, err := headerLink.Parse(linkText[0])

		if err != nil {
			return nil, err
		}

		linkMap := map[string]headerLink.Link{}
		for _, link := range links {
			linkMap[link.Rel] = link
		}

		return linkMap, nil
	}

	return map[string]headerLink.Link{}, nil
}

// populateRate parses the rate related headers and populates the response Rate.
func (r *Response) populateRate() {
	if limit := r.Header.Get(headerRateLimit); limit != "" {
		r.Rate.Limit, _ = strconv.Atoi(limit)
	}
	if remaining := r.Header.Get(headerRateRemaining); remaining != "" {
		r.Rate.Remaining, _ = strconv.Atoi(remaining)
	}
	if reset := r.Header.Get(headerRateReset); reset != "" {
		if v, _ := strconv.ParseInt(reset, 10, 64); v != 0 {
			r.Rate.Reset = Timestamp{time.Unix(v, 0)}
		}
	}
}

// Do sends an API request and returns the API response. The API response is JSON decoded and stored in the value
// pointed to by v, or returned as an error if an API error has occurred. If v implements the io.Writer interface,
// the raw response will be written to v, without attempting to decode it.
func (c *Client) Do(req *http.Request, v interface{}) (*Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	response := newResponse(resp)
	c.Rate = response.Rate

	err = CheckResponse(resp)
	if err != nil {
		return response, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			json.NewDecoder(resp.Body).Decode(v)
		}
	}

	return response, err
}
func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message)
}

// CheckResponse checks the API response for errors, and returns them if present. A response is considered an
// error if it has a status code outside the 200 range. API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse. Any other response body will be silently ignored.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		json.Unmarshal(data, errorResponse)
	}

	return errorResponse
}

func (r Rate) String() string {
	return Stringify(r)
}

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string {
	p := new(string)
	*p = v
	return p
}

// Int is a helper routine that allocates a new int32 value
// to store v and returns a pointer to it, but unlike Int32
// its argument value is an int.
func Int(v int) *int {
	p := new(int)
	*p = v
	return p
}

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool {
	p := new(bool)
	*p = v
	return p
}

// StreamToString converts a reader to a string
func StreamToString(stream io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.String()
}