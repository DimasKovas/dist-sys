FROM    golang:latest
WORKDIR /go/src/item-storage
COPY ./wait-for-it/wait-for-it.sh .
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build
