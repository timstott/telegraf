package transmission

// Transmission RPC inspired from https://github.com/odwrtw/transmission
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type rpcClient struct {
	Address    *url.URL
	HTTPClient *http.Client
	sessionID  string
}

type Tracker struct {
	Announce string
}

type torrent struct {
	DownloadedEver int64
	Error          int
	HashString     string
	Name           string
	PeersConnected int
	PercentDone    float64
	RateDownload   int
	RateUpload     int
	Status         int
	Trackers       []Tracker
	UploadedEver   int64
}

type torrents struct {
	Torrents []*torrent `json:"torrents"`
}

// Request object for API call
type request struct {
	Method    string      `json:"method"`
	Arguments interface{} `json:"arguments"`
}

// Response object for API call response
type response struct {
	Arguments interface{} `json:"arguments"`
	Result    string      `json:"result"`
}

func (c *rpcClient) makeRequest(tReq *request, tResp *response) error {
	data, err := json.Marshal(tReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.Address.String(), bytes.NewReader(data))
	if err != nil {
		return err
	}

	resp, err := c.Do(req, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("got http error %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, tResp)
	if err != nil {
		return err
	}

	if tResp.Result != "success" {
		return fmt.Errorf("transmission: request response %q", tResp.Result)
	}
	return nil
}

func (c *rpcClient) Do(req *http.Request, retry bool) (*http.Response, error) {
	if c.sessionID != "" {
		// Delete the previous session in header in case there was one

		// We need to do this because the previous header won't be overridden
		// by the "Add", we need to manually delete the previous header or the
		// request will fail
		req.Header.Del("X-Transmission-Session-Id")
		req.Header.Add("X-Transmission-Session-Id", c.sessionID)
	}

	// Body copy for replay it if needed
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body.Close()
	req.Body = ioutil.NopCloser(bytes.NewBuffer(b))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Most Transmission RPC servers require a X-Transmission-Session-Id header
	// to be sent with requests, to prevent CSRF attacks.  When your request
	// has the wrong id -- such as when you send your first request, or when
	// the server expires the CSRF token -- the Transmission RPC server will
	// return an HTTP 409 error with the right X-Transmission-Session-Id in its
	// own headers.  So, the correct way to handle a 409 response is to update
	// your X-Transmission-Session-Id and to resend the previous request.
	if resp.StatusCode == http.StatusConflict && retry {
		c.sessionID = resp.Header.Get("X-Transmission-Session-Id")

		// Copy the previous request body in order to do it again
		req.Body = ioutil.NopCloser(bytes.NewBuffer(b))

		// We also need to read the body before closing it, or it will trigger
		// a "net/http: request cancelled" error
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()

		return c.Do(req, false)
	}

	return resp, nil
}
func (c *rpcClient) getTorrents() ([]*torrent, error) {
	type arg struct {
		Fields []string `json:"fields"`
	}

	tReq := &request{
		Arguments: arg{
			Fields: []string{
				"downloadedEver",
				"error",
				"hashString",
				"name",
				"peersConnected",
				"percentDone",
				"rateDownload",
				"rateUpload",
				"status",
				"trackers",
				"uploadedEver",
			},
		},
		Method: "torrent-get",
	}

	r := &response{Arguments: &torrents{}}

	err := c.makeRequest(tReq, r)
	if err != nil {
		return nil, err
	}
	return r.Arguments.(*torrents).Torrents, nil
}
