FROM    golang:latest
WORKDIR /go/src/provider-mock
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build
