package dbclient

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
)

type Client struct {
	connection *pgx.Conn
}

var ErrNotFound = errors.New("Not found found")
var ErrUserAlreadyExists = errors.New("User with this username already exists")
var ErrTokenAlreadyExists = errors.New("")

func isDuplicateError(err error) bool {
	return strings.Contains(fmt.Sprint(err), "duplicate key value violates unique constraint")
}

func CreateDbClient() (Client, error) {
	db := Client{}
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return Client{}, err
	}
	db.connection = conn
	_, err = db.connection.Exec(context.Background(),
		`create table if not exists users (
			username text primary key,
			pass_hash text not null,
			email text not null
		);`)
	if err != nil {
		return Client{}, err
	}
	_, err = db.connection.Exec(context.Background(),
		`create table if not exists tokens (
			token text primary key,
			exp_time timestamp not null,
			refresh boolean not null,
			username text not null
		);`)
	return db, err
}

type User struct {
	Username string
	PassHash string
	Email    string
}

func (db *Client) AddUser(user User) error {
	_, err := db.connection.Exec(context.Background(),
		`insert into users (username, pass_hash, email)
		values ($1, $2, $3)
		`, user.Username, user.PassHash, user.Email)
	if isDuplicateError(err) {
		return ErrUserAlreadyExists
	}
	return err
}

func (db *Client) GetUser(username string) (User, error) {
	var user User
	err := db.connection.QueryRow(context.Background(),
		`select username, pass_hash, email from users
		where username = $1
		`, username).Scan(&user.Username, &user.PassHash, &user.Email)
	if err == pgx.ErrNoRows {
		err = ErrNotFound
	}
	return user, err
}

type TokenInfo struct {
	Token    string
	ExpTime  time.Time
	Refresh  bool
	Username string
}

func (db *Client) GetTokenInfo(token string) (TokenInfo, error) {
	var tinfo TokenInfo
	err := db.connection.QueryRow(context.Background(),
		`select token, exp_time, refresh, username from tokens
		where token = $1
		`, token).Scan(&tinfo.Token, &tinfo.ExpTime, &tinfo.Refresh, &tinfo.Username)
	if err == pgx.ErrNoRows {
		err = ErrNotFound
	}
	return tinfo, err
}

func (db *Client) AddNewToken(tinfo TokenInfo) error {
	_, err := db.connection.Exec(context.Background(),
		`insert into tokens (token, exp_time, refresh, username)
		values($1, $2, $3, $4)`, tinfo.Token, tinfo.ExpTime, tinfo.Refresh, tinfo.Username)
	if isDuplicateError(err) {
		return ErrTokenAlreadyExists
	}
	return err
}
