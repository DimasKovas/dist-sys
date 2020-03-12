package dbclient

import (
	"context"
	"os"

	"github.com/jackc/pgx"
)

type Client struct {
	connection *pgx.Conn
}

func CreateDbClient() (Client, error) {
	db := Client{}
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return Client{}, err
	}
	db.connection = conn
	_, err = db.connection.Exec(context.Background(),
		`create table if not exists items (
			id serial primary key,
			title text not null,
			category text not null
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
	_, err := db.connection.Exec(context.Background(),
		`update items
		set title = $1,
			category = $2
		where id = $3
		`, item.Title, item.Category, item.ID)
	return err
}

func (db *Client) DeleteItem(id uint64) error {
	_, err := db.connection.Exec(context.Background(),
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
	Offset     uint64
	Limit      uint64
	Categories []string // Unsupported yet
}

func (db *Client) GetItemList(options GetItemListOptions) ([]Item, error) {
	query := `select id, title, category from items
		order by id
		offset $1
		limit $2`
	rows, err := db.connection.Query(context.Background(), query,
		options.Offset, options.Limit)
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
