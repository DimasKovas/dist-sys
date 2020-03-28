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
			phone_number text not null,
			phone_confirmed boolean not null
		);`)
	if err != nil {
		return Client{}, err
	}
	_, err = db.connection.Exec(context.Background(),
		`create table if not exists tokens (
			token text primary key,
			exp_time timestamp not null,
			token_type integer not null,
			username text not null
		);`)
	return db, err
}

type User struct {
	Username       string
	PassHash       string
	PhoneNumber    string
	PhoneConfirmed bool
}

func (db *Client) AddUser(user User) error {
	_, err := db.connection.Exec(context.Background(),
		`insert into users (username, pass_hash, phone_number, phone_confirmed)
		values ($1, $2, $3, $4)
		`, user.Username, user.PassHash, user.PhoneNumber, user.PhoneConfirmed)
	if isDuplicateError(err) {
		return ErrUserAlreadyExists
	}
	return err
}

func (db *Client) GetUser(username string) (User, error) {
	var user User
	err := db.connection.QueryRow(context.Background(),
		`select username, pass_hash, phone_number, phone_confirmed from users
		where username = $1
		`, username).Scan(&user.Username, &user.PassHash, &user.PhoneNumber, &user.PhoneConfirmed)
	if err == pgx.ErrNoRows {
		err = ErrNotFound
	}
	return user, err
}

func (db *Client) ConfirmPhoneNumber(username string) error {
	tags, err := db.connection.Exec(context.Background(),
		`update users
		set phone_confirmed = true
		where username = $1
		`, username)
	if err == nil && tags.RowsAffected() != 1 {
		err = ErrNotFound
	}
	return err
}

type TokenType int

const (
	ACCESS  TokenType = 0
	REFRESH TokenType = 1
	CONFIRM TokenType = 2
)

type TokenInfo struct {
	Token    string
	ExpTime  time.Time
	Type     TokenType
	Username string
}

func (db *Client) GetTokenInfo(token string) (TokenInfo, error) {
	var tinfo TokenInfo
	err := db.connection.QueryRow(context.Background(),
		`select token, exp_time, token_type, username from tokens
		where token = $1
		`, token).Scan(&tinfo.Token, &tinfo.ExpTime, &tinfo.Type, &tinfo.Username)
	if err == pgx.ErrNoRows {
		err = ErrNotFound
	}
	return tinfo, err
}

func (db *Client) AddNewToken(tinfo TokenInfo) error {
	_, err := db.connection.Exec(context.Background(),
		`insert into tokens (token, exp_time, token_type, username)
		values($1, $2, $3, $4)`, tinfo.Token, tinfo.ExpTime, tinfo.Type, tinfo.Username)
	if isDuplicateError(err) {
		return ErrTokenAlreadyExists
	}
	return err
}
