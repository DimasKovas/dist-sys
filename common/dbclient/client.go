package dbclient

import (
	"context"
	"errors"
	"hash/crc64"
	"os"
	"strconv"

	"github.com/jackc/pgx/v4"
)

type Client struct {
	connection *pgx.Conn
}

var ErrNotFound = errors.New("Specified item doesn't exist")

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
			universal_code text unique not null,
			title text not null,
			category text not null
		);`)
	return db, err
}

type Item struct {
	ID            uint64 `json:"id"`
	UniversalCode string `json:"universal_code"`
	Title         string `json:"title"`
	Category      string `json:"category"`
}

func (item *Item) GenerateUniversalCode() {
	table := crc64.MakeTable(123)
	item.UniversalCode = strconv.FormatUint(
		(crc64.Checksum([]byte(item.Title), table)*399123321)^
			crc64.Checksum([]byte(item.Category), table),
		16)
}

func (db *Client) NewItem(item Item) (uint64, error) {
	if item.UniversalCode == "" {
		item.GenerateUniversalCode()
	}
	var id uint64
	err := db.connection.QueryRow(context.Background(),
		`insert into items (universal_code, title, category)
		values ($1, $2, $3)
		returning id;
		`, item.UniversalCode, item.Title, item.Category).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (db *Client) UpdateItem(item Item) error {
	if item.UniversalCode == "" {
		item.GenerateUniversalCode()
	}
	tags, err := db.connection.Exec(context.Background(),
		`update items
		set universal_code = $1
			title = $2
			category = $3
		where id = $4
		`, item.UniversalCode, item.Title, item.Category, item.ID)
	if err == nil && tags.RowsAffected() != 1 {
		return ErrNotFound
	}
	return err
}

func (db *Client) DeleteItem(id uint64) error {
	tags, err := db.connection.Exec(context.Background(),
		`delete from items
		where id = $1;
		`, id)
	if err == nil && tags.RowsAffected() != 1 {
		return ErrNotFound
	}
	return err
}

func (db *Client) GetItem(id uint64) (Item, error) {
	var item Item
	err := db.connection.QueryRow(context.Background(),
		`select id, universal_code, title, category from items where id = $1
		`, id).Scan(&item.ID, &item.UniversalCode, &item.Title, &item.Category)
	if err == pgx.ErrNoRows {
		err = ErrNotFound
	}
	return item, err
}

type GetItemListOptions struct {
	Offset     uint64
	Limit      uint64
	Categories []string // Unsupported yet
}

func (db *Client) GetItemList(options GetItemListOptions) ([]Item, error) {
	query := `select id, universal_code, title, category from items
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
		err := rows.Scan(&item.ID, &item.UniversalCode, &item.Title, &item.Category)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (db *Client) GetItemListSize() (int, error) {
	var size int
	err := db.connection.QueryRow(context.Background(),
		`select count(*) from items`).Scan(&size)
	return size, err
}

func (db *Client) ImportItemBatch(batch []Item) error {
	dbbatch := &pgx.Batch{}
	query := `insert into items (universal_code, title, category)
		values ($1, $2, $3) ON CONFLICT DO NOTHING`
	for _, item := range batch {
		if item.UniversalCode == "" {
			item.GenerateUniversalCode()
		}
		dbbatch.Queue(query, item.UniversalCode, item.Title, item.Category)
	}
	batch_results := db.connection.SendBatch(context.Background(), dbbatch)
	defer batch_results.Close()
	_, err := batch_results.Exec()
	return err
}
