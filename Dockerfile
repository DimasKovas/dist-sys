FROM    golang:latest
WORKDIR /go/src/github.com/DimasKovas/dist-sys/
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN git submodule update --init --recursive
RUN go build main.go
CMD  ["./main"]
