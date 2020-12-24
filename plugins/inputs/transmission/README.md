# Example Input Plugin

Read metrics from Transmission server via the RPC interface.

Telegraf minimum version: Telegraf x.x
Transmission minimum tested version: 3.0

### TODO Configuration

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage <plugin-name>`.

```toml
[[inputs.example]]
  example_option = "example_value"
```

#### example_option

A more in depth description of an option can be provided here, but only do so
if the option cannot be fully described in the sample config.

### Metrics

- torrents
  - tags:
    - name
    - hash_string
    - tracker # TODO: torrent with multiple trackers will cause duplicate fields
    - error
    - status
  - fields:
    - downloaded
    - peer_connected
    - percent_done
    - rate_download (B/s)
    - rate_upload (B/s)
    - uploaded

### Sample Queries

This section can contain some useful InfluxDB queries that can be used to get
started with the plugin or to generate dashboards.  For each query listed,
describe at a high level what data is returned.

Get the max, mean, and min for the measurement in the last hour:
```
SELECT max(field1), mean(field1), min(field1) FROM measurement1 WHERE tag1=bar AND time > now() - 1h GROUP BY tag
```

### Example Output

This section shows example output in Line Protocol format.  You can often use
`telegraf --input-filter <plugin-name> --test` or use the `file` output to get
this information.

```
measurement1,tag1=foo,tag2=bar field1=1i,field2=2.1 1453831884664956455
measurement2,tag1=foo,tag2=bar,tag3=baz field3=1i 1453831884664956455
```
