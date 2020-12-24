# Example Input Plugin

Read metrics from Transmission server via the RPC interface.

Telegraf minimum version: Telegraf 1.17
Transmission minimum tested version: 3.0

### Configuration

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage transmission`.

```toml
# Read Transmission torrent metrics via RPC interface
[[inputs.transmission]]
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
```

### Metrics

- transmission
  - fields:
    - downloaded
    - error
    - peer_connected
    - percent_done
    - rate_download (B/s)
    - rate_upload (B/s)
    - uploaded
  - tags:
    - name
    - hash
    - status
    - tracker (Only takes first tracker in the list)
    - host
    - port

### Example Output

```
measurement1,tag1=foo,tag2=bar field1=1i,field2=2.1 1453831884664956455
measurement2,tag1=foo,tag2=bar,tag3=baz field3=1i 1453831884664956455
```
