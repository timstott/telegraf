package transmission

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const rpcGetTorrentResponse = `
{
  "arguments": {
    "torrents": [
      {
        "downloadedEver": 1000,
        "error": 0,
        "hashString": "000000",
        "name": "Macron Super Star",
        "peersConnected": 789,
        "percentDone": 0.75,
        "rateDownload": 2000,
        "rateUpload": 4000,
        "status": 6,
        "trackers": [
          {
            "announce": "http://tracker.hello.cc:2710/abc/announce",
            "id": 0,
            "scrape": "http://tracker.hello.cc:2710/abc/scrape",
            "tier": 0
          }
        ],
        "uploadedEver": 222
      }
    ]
  },
  "result": "success"
}
`

func TestTransmissionGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverSessionID := "session-id-123"
		requestSessionID := r.Header.Get("X-Transmission-Session-Id")
		var rsp string

		if requestSessionID != serverSessionID {
			w.Header().Set("X-Transmission-Session-Id", serverSessionID)
			w.WriteHeader(http.StatusConflict)

			return
		}

		if r.URL.Path == "/transmission/rpc" {
			rsp = rpcGetTorrentResponse
		} else {
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	plugin := &Transmission{
		Url: fmt.Sprintf("%s/transmission/rpc", ts.URL),
	}

	var acc_transmission testutil.Accumulator

	err_transmission := acc_transmission.GatherError(plugin.Gather)

	require.NoError(t, err_transmission)

	expected_fields := map[string]interface{}{
		"downloaded":      int64(1000),
		"error":           int(0),
		"peers_connected": int(789),
		"uploaded":        int64(222),
		"rate_download":   int(2000),
		"rate_upload":     int(4000),
		"percent_done":    float64(0.75),
	}

	addr, err := url.Parse(ts.URL)
	if err != nil {
		panic(err)
	}
	host, port, _ := net.SplitHostPort(addr.Host)

	expected_tags := map[string]string{
		"hash":    "000000",
		"host":    host,
		"name":    "Macron Super Star",
		"port":    port,
		"status":  "seeding",
		"tracker": "tracker.hello.cc",
	}

	acc_transmission.AssertContainsTaggedFields(
		t,
		"transmission",
		expected_fields,
		expected_tags,
	)
}
