package transmission

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Transmission struct {
	Url             string
	ResponseTimeout internal.Duration
	tls.ClientConfig

	// Transmission RPC client
	client *rpcClient
}

var sampleConfig = `
  # Transmission RPC interface URL.
  url = "http://localhost:9091/transmission/rpc"

  ## Optional TLS Config
  tls_ca = "/etc/telegraf/ca.pem"
  tls_cert = "/etc/telegraf/cert.cer"
  tls_key = "/etc/telegraf/key.key"
  ## Use TLS but skip chain & host verification
  insecure_skip_verify = false

  # HTTP response timeout (default: 5s)
  response_timeout = "5s"
`

func (n *Transmission) SampleConfig() string {
	return sampleConfig
}

func (n *Transmission) Description() string {
	return "Read Transmission torrent metrics via RPC interface"
}

func (n *Transmission) Gather(acc telegraf.Accumulator) error {
	address, err := url.Parse(n.Url)
	if err != nil {
		acc.AddError(fmt.Errorf("Unable to parse address '%s': %s", n.Url, err))
		return nil
	}

	// Create one Transmission RPC client that is re-used for each
	// collection interval
	if n.client == nil {
		httpClient, err := n.createHttpClient()
		if err != nil {
			return err
		}

		n.client = &rpcClient{
			Address:    address,
			HTTPClient: httpClient,
		}
	}

	acc.AddError(n.gatherMetrics(acc))

	return nil
}

func (n *Transmission) createHttpClient() (*http.Client, error) {
	tlsCfg, err := n.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	if n.ResponseTimeout.Duration < time.Second {
		n.ResponseTimeout.Duration = time.Second * 5
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: n.ResponseTimeout.Duration,
	}

	return client, nil
}

func (n *Transmission) gatherMetrics(acc telegraf.Accumulator) error {
	torrents, err := n.client.getTorrents()
	if err != nil {
		return err
	}

	host, port := splitHostPort(n.client.Address)

	for _, tor := range torrents {
		fields := map[string]interface{}{
			"downloaded":      tor.DownloadedEver,
			"error":           tor.Error,
			"peers_connected": tor.PeersConnected,
			"rate_download":   tor.RateDownload,
			"rate_upload":     tor.RateUpload,
			"uploaded":        tor.UploadedEver,
			"percent_done":    tor.PercentDone,
		}

		tags := map[string]string{
			"hash":    tor.HashString,
			"host":    host,
			"name":    tor.Name,
			"port":    port,
			"status":  getTorrentStatus(tor.Status),
			"tracker": getTorrentTracker(tor.Trackers),
		}

		acc.AddFields("transmission", fields, tags)
	}

	return nil
}

func getTorrentTracker(trackers []Tracker) string {
	if len(trackers) == 0 {
		return "unknown"
	}

	address, err := url.Parse(trackers[0].Announce)

	if err != nil {
		return "parse_error"
	}

	return address.Hostname()
}

// Humanize ...
// https://github.com/transmission/transmission/blob/2c7a9999022050d2a2ac47ac2ad430fe52e49a4d/libtransmission/transmission.h#L1660
func getTorrentStatus(statusCode int) string {
	statuses := map[int]string{
		0: "stopped",
		1: "check_wait",
		2: "checking",
		3: "download_wait",
		4: "downloading",
		5: "seed_wait",
		6: "seeding",
	}

	status, ok := statuses[statusCode]

	if !ok {
		status = "unknown"
	}

	return status
}

func splitHostPort(addr *url.URL) (string, string) {
	h := addr.Host
	host, port, err := net.SplitHostPort(h)
	if err != nil {
		host = addr.Host
		if addr.Scheme == "http" {
			port = "80"
		} else if addr.Scheme == "https" {
			port = "443"
		} else {
			port = ""
		}
	}
	return host, port
}

func init() {
	inputs.Add("transmission", func() telegraf.Input {
		return &Transmission{}
	})
}
