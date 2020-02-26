FROM    golang:latest
WORKDIR /usr/src/app/
COPY    . .
RUN go build main.go
CMD  ["./main"]