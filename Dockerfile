FROM golang:1.27 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -v \
    -ldflags "-s -w -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) -X main.gitCommit=${GIT_COMMIT:-unknown}" \
    -o /todoapp ./cmd/todoapp

FROM scratch
COPY --from=build /todoapp /todoapp
VOLUME /data
EXPOSE 8080
ENTRYPOINT ["/todoapp", "--data-dir", "/data", "--addr", ":8080"]
