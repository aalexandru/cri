FROM golang:1.17
WORKDIR ${GOPATH}/src/github.com/aalexandru/cri
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /bin/cri main.go

FROM alpine
COPY --from=0 /bin/cri /usr/local/bin/