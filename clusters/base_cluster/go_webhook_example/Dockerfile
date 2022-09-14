FROM golang:1.19 AS build

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /work
COPY . /work

# Build mutation-webhook
RUN --mount=type=cache,target=/root/.cache/go-build,sharing=private \
    go build -o src/mutation-webhook .

# ---
FROM scratch AS run

COPY --from=build /work/bin/ /usr/local/bin/

CMD ["mutation-webhook"]