package main

import (
	"context"
	"errors"

	"cloud.google.com/go/datastore"
)

// Errors
var (
	ErrNoUserFound = errors.New("database: no user found")
)

type Database struct {
	ds *datastore.Client
}

func initDB() (*Database, error) {
	ctx := context.Background()

	client, err := datastore.NewClient(ctx, *googleProjectID)
	if err != nil {
		return nil, err
	}

	db := &Database{
		ds: client,
	}

	return db, err
}

func (d *Database) GetUser(ctx context.Context, id string) (*DiscordUser, *datastore.Key, error) {
	result := &DiscordUser{}
	var resultKey *datastore.Key

	q := datastore.NewQuery("DiscordUser").Filter("ID =", id).Limit(1)

	for t := d.ds.Run(ctx, q); ; {
		val := &DiscordUser{}

		key, err := t.Next(val)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, nil, err
		}

		result = val
		resultKey = key
	}

	if len(result.ID) == 0 {
		return nil, nil, ErrNoUserFound
	}

	return result, resultKey, nil
}

func (d *Database) PutUser(ctx context.Context, dUser *DiscordUser) (*datastore.Key, error) {
	_, _, err := d.GetUser(ctx, dUser.ID)
	if err != ErrNoUserFound {
		return nil, err
	}

	k := datastore.NewIncompleteKey(ctx, "DiscordUser", nil)
	return d.ds.Put(ctx, k, dUser)
}
