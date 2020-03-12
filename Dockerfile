FROM    golang:latest
WORKDIR /go/src/github.com/DimasKovas/dist-sys/
COPY    . .
RUN go get -u
RUN go build main.go
RUN wget https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh
RUN chmod 0777 wait-for-it.sh
CMD  ["./main"]