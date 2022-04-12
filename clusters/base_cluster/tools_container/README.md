The tools container includes things like dnsutils and curl.

Build and run locally:
```
docker build . -t dnsutils
docker run --rm -it --entrypoint bash dnsutils
```