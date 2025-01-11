# prometheus-dashboard

A simple web server to serve and show some Prometheus metrics and Alertmanager alerts

<img src="screenshot.png" alt="screenshot" width="680"/>

This tool is for people who has running Prometheus and Alertmanager in the backend and want to show some few metrics on the web without heavy tools like Grafana. Use case is a kind of status page or status view on mobile app-

## precondition

- Prometheus running on http://127.0.0.1:9090
- Alertmanager running on http://127.0.0.1:9093
- adjust the metrics in the index.html template

## usage

The dashboard will serve on 0.0.0.0:8080 and can be visit by http://<your-install-server>:8080

### use as single binary, downloaded from the release page:

```
./prometheus-dashboard
```

### use as Docker image and run it:

```
docker run -d -P 8080:8080 ghcr.io/eumel8/charts/prometheus-dashboard:1.0.4
```

set another Prometheus and Alertmanager URL

```
docker run -d -P 8080:8080 -e PROMETHEUS_URL=http://192.168.0.49:9090/api/v1/query -e ALERTMANAGER_URL=http://192.168.0.49:9093/api/v2/alerts ghcr.io/eumel8/charts/prometheus-dashboard:1.0.4
```

hint: with `docker -v` you can mount another index.html template in your docker

### use with Helm in Kubernetes

```
helm -n monitoring upgrade dashboard oci://ghcr.io/eumel8/charts/prometheus-dashboard:1.0.4 --create-namespace
```

set another Prometheus and Alertmanager URL

```
helm -n monitoring upgrade dashboard oci://ghcr.io/eumel8/charts/prometheus-dashboard:1.0.4 --create-namespace --set prometheusURL="http://92.168.0.49:9090/api/v1/query" --set alertmanagerURL="http://192.168.0.49:9093/api/v2/alerts"
```

hint: you can manage your own index.html template as ConfigMap and refer this in the Helm chart values:

```
indexMap:
  enabled: true
  name:  my-index-html
```

## Thanos

You can also query a Thanos backend instead of Prometheus. Use the `-thanos` flag or the Helm chart option:

```
thanos:
  enabled: true
```

Move also the thanos.html template to index.html to handle the different data struct.

## Credits

Frank Kloeker <f.kloeker@telekom.de>

Life is for sharing. If you have an issue with the code or want to improve it,
feel free to open an issue or an pull request.
