# Glog, a cli for Google Cloud Logging

I don't like viewing logs via web UIs, so i built this tool to be able to easily tail logs from my terminal.

### Some examples:

Tail logs for service.foo
```
glog -f -p my-project -q 'resource.labels.service_name = "service.foo"'
```
Get all logs from your project for the last 24h
```
glog -d 24h -p my-project
```

You can learn more about how to build queries for the -q parameter at https://cloud.google.com/logging/docs/view/query-library-preview?hl=en-GB

#### Usage: 
```
glog -help 
  -d string
        lookback duration, eg '1h' or '30s' (default "1h")
  -f    follow the logs, like tail -f
  -p string
        your project ID
  -q string
        log query

```
