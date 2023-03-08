# Grafana Warp10 Datasource Backend Plugin

This plugin is not official and is not maintained by [SenX](https://senx.io/).

The purpose of this plugin is to be able to use Grafana alerts with [Warp10](https://warp10.io/).

## Plugin

### Configuration

- `Base url`: URL of Warp10 instance, the plugin will add Warp10 api path (`/api/v0/{fetch,find,...}`).
- `Read token`: Warp10 token can be used with variable `$read_token` in query. The plugin will check if the token is valid and is a read token.

### Query

Currently, only JSON time serie format is supported (GTS format with array of two elements for the value).

[More information on Warp10 GTS format.](https://www.warp10.io/content/03_Documentation/03_Interacting_with_Warp_10/04_Fetching_data/02_GTS_JSON_output_format)

### Query variables

To simplify query this plugin provide variables:

- `$read_token`: Read token configured in Datasource.
- `$fromISO`: ISO datetime string of the start of the currently active date.
- `$toISO`: ISO datetime string of the end of the currently active date.
- `$interval`: Microseconds value of the Grafana's calculated interval.

Example:

```warpscript
[
  $read_token
  'telegraf.mem.used_percent'
  {  }
  $toISO
  $fromISO
] FETCH

[ SWAP bucketizer.mean 0 $interval 0 ] BUCKETIZE
```

### Query options

- `Show labels`: Show labels of the Warp10 GTS.
- `Pattern namming`: Pattern to rename serie. Pattern `$label_<warp10_label_name>` will be interpolated with the value of the Warp10 label if it exists.
