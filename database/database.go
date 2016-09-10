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
	tx, err := d.ds.NewTransaction(d.ctx, datastore.MaxAttempts(1))
	if err != nil {
		return nil, err
	}

	dUser, _, err := d.getUser(d.ctx, tx, id)
	if err != nil {
		return nil, err
	}

	_, err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return dUser, err
}

func (d *Database) getUser(ctx context.Context, tx *datastore.Transaction, id string) (*interop.DiscordUser, *datastore.Key, error) {
	results := []interop.DiscordUser{}
	var result *interop.DiscordUser
	var resultKey *datastore.Key

	q := datastore.NewQuery("DiscordUser").
		Filter("ID =", id).
		Limit(1).
		Transaction(tx)

	resultKeys, err := d.ds.GetAll(ctx, q, &results)
	if err != nil {
		switch err {
		case datastore.ErrNoSuchEntity:
			return nil, nil, ErrNoUserFound
		default:
			return nil, nil, err
		}
	}

	result = &results[0]
	resultKey = resultKeys[0]

	if len(result.ID) == 0 {
		return nil, nil, ErrNoUserFound
	}

	return result, resultKey, nil
}

func (d *Database) PutUser(dUser *interop.DiscordUser) error {
	tx, err := d.ds.NewTransaction(d.ctx, datastore.MaxAttempts(1))
	if err != nil {
		return err
	}

	_, err = d.putUser(d.ctx, tx, dUser)
	if err != nil {
		return err
	}

	_, err = tx.Commit()
	if err != nil {
		return err
	}

	return err
}

func (d *Database) putUser(ctx context.Context, tx *datastore.Transaction, dUser *interop.DiscordUser) (*datastore.PendingKey, error) {
	_, key, err := d.getUser(ctx, tx, dUser.ID)
	if err != nil && err != ErrNoUserFound {
		tx.Rollback()
		return nil, err
	}

	if key == nil {
		key = datastore.NewIncompleteKey(ctx, "DiscordUser", nil)
	}

	return tx.Put(key, dUser)
}
