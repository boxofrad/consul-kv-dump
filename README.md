# `consul-kv-dump`

Quick and dirty tool for dumping all of the KV data in a Consul snapshot to a
JSON file, for ingestion into another system. Useful in scenarios where you
don't want to disrupt production traffic by running `consul kv export`.

## Usage

```shell
$ consul snapshot save backup.snap
$ consul-kv-dump backup.snap dump.json
```
