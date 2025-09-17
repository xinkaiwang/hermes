# How to build?
```
make
make docker-build
make docker-push
```

# How to post events?

```
curl -k http://localhost:8080/api/post -d '{"events": [{"modle": "loader.routeMiddleware", "event":"AddMiddleware", "type":"UserData", "route":"/delay", "service":"rain"}]}'
curl -k http://localhost:8080/api/post -d '{"events": [{"modle": "loader.routeMiddleware", "event":"AddMiddleware", "type":"UserData", "route":"/delay", "service":"rain", "host":"127.0.0.1"},{"event":"AddMiddleware", "type":"UserData2", "route":"/delay2", "service":"rain2", "host":"127.0.0.2"}]}'
```
