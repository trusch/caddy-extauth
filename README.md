# caddy-extauth
an external authentication plugin for caddy

## Quickstart

### Requirements

You need `podman` and `buildah` to use the projects Makefile. If you don't have those tools and don't want to install them, just go with the `xcaddy` instructions and use this repo as plugin module.

### Build and Run
```bash
> make all
> make run
> curl -H "Authorization: secret" localhost:2015
You are authenticated!
> podman logs caddy-extauth-auth
{"level":"info","ip":"127.0.0.1:33554","time":"2020-10-16T06:09:33Z","message":"success"}
> podman logs caddy-extauth-logger
---
GET / HTTP/1.1
Host: localhost:2015
User-Agent: curl/7.72.0
Accept: */*
Accept-Encoding: gzip
Authorization: secret
X-Forwarded-For: 127.0.0.1
X-Forwarded-Proto: http
X-Token: token
> make stop
```

There is also a compose.yaml file for your reference.
