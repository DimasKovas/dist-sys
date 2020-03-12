package dbclient

import (
	"context"
	"os"

	"github.com/jackc/pgx/v4"
)

type Client struct {
	connection *pgx.Conn
}

func CreateDbClient() (DbClient, error) {
	db := Client{}
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return Client{}, err
	}
	db.connection = conn
	_, err = db.connection.Exec(context.Background(),
		`create table if not exists items (
			id serial primary key,
			title string not null,
			category string not null
		);`)
	if err != nil {
		return Client{}, err
	}
	return db, nil
}

type Item struct {
	ID       uint64 `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
}

func (db *Client) NewItem(item Item) (uint64, error) {
	var id uint64
	err := db.connection.QueryRow(context.Background(),
		`insert into items (title, category)
		values ($1, $2)
		returning id;
		`, item.Title, item.Category).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (db *Client) UpdateItem(item Item) error {
	err := db.connection.QueryRow(context.Background(),
		`update items
		set title = $1,
			category = $2
		where id = $3
		`, item.Title, item.Category, item.ID)
	return err
}

func (db *Client) DeleteItem(id uint64) error {
	err := db.connection.QueryRow(context.Background(),
		`delete from items
		where id = $1;
		`, id)
	return err
}

func (db *Client) GetItem(id uint64) (Item, error) {
	var item Item
	err := db.connection.QueryRow(context.Background(),
		`select id, title, category from items where id = $1
		`, id).Scan(&item.ID, &item.Title, &item.Category)
	if err != nil {
		return Item{}, err
	}
	return item, nil
}

type GetItemListOptions struct {
	Offset     *uint
	Limit      *uint
	Categories []string
}

func (db *Client) GetItemList(options GetItemListOptions) ([]Item, error) {
	query := "select id, title, category from items ordered by id"
	var offset, limit uint
	var categories []string
	if options.category != nil {
		query += " where category in $3"
		categories = options.Categories
	}
	if options.Offset != nil {
		query += " offset $1"
		offset = *options.Offset
	}
	if options.Limit != nil {
		query += " limit $2"
		limit = *options.Limit
	}
	rows, err := db.connection.Query(context.Background(), query, offset, limit, categories)
	if err != nil {
		return nil, err
	}
	res := make([]Item, 0)
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Title, &item.Category)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}
