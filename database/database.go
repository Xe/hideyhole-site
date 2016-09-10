package database

import (
	"context"
	"errors"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/Xe/hideyhole-site/interop"
)

// Errors
var (
	ErrNoUserFound = errors.New("database: no user found")
)

type Database struct {
	ds  *datastore.Client
	ctx context.Context
}

//hack
func init() {
	http.DefaultServeMux = http.NewServeMux()
}

func Init(namespace, projectID string) (*Database, error) {
	ctx := context.Background()

	ctx = datastore.WithNamespace(ctx, namespace)

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	db := &Database{
		ds:  client,
		ctx: ctx,
	}

	return db, err
}

func (d *Database) GetUser(id string) (*interop.DiscordUser, error) {
	dUser, _, err := d.getUser(d.ctx, id)

	return dUser, err
}

func (d *Database) getUser(ctx context.Context, id string) (*interop.DiscordUser, *datastore.Key, error) {
	result := &interop.DiscordUser{}
	var resultKey *datastore.Key

	q := datastore.NewQuery("DiscordUser").Filter("ID =", id).Limit(1)

	for t := d.ds.Run(d.ctx, q); ; {
		val := &interop.DiscordUser{}

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

func (d *Database) PutUser(dUser *interop.DiscordUser) error {
	_, err := d.putUser(d.ctx, dUser)

	return err
}

func (d *Database) putUser(ctx context.Context, dUser *interop.DiscordUser) (*datastore.Key, error) {
	_, key, err := d.getUser(ctx, dUser.ID)
	if err != nil && err != ErrNoUserFound {
		return nil, err
	}

	if key == nil {
		key = datastore.NewIncompleteKey(ctx, "DiscordUser", nil)
	}

	return d.ds.Put(ctx, key, dUser)
}
