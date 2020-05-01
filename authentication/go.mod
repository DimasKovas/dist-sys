module authentication

go 1.13

require (
	common v0.0.0-00010101000000-000000000000
	github.com/jackc/pgx/v4 v4.6.0
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	golang.org/x/crypto v0.0.0-20200323165209-0ec3e9974c59
	google.golang.org/grpc v1.29.1
)

replace common => ../common
