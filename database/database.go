package database

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Xe/hideyhole-site/interop"
	"github.com/bwmarrin/snowflake"
)

// Errors
var (
	ErrNoUserFound = errors.New("database: no user found")
	ErrNoFicFound  = errors.New("database: no fic found")
)

type Database struct {
	ds    *datastore.Client
	ctx   context.Context
	IDGen *snowflake.Node
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

	node, err := snowflake.NewNode(rand.Int63() % 1024)
	if err != nil {
		return nil, err
	}

	db := &Database{
		ds:    client,
		ctx:   ctx,
		IDGen: node,
	}

	return db, err
}

func (d *Database) GetUser(id string) (*interop.DiscordUser, error) {
	tx, err := d.ds.NewTransaction(d.ctx, datastore.MaxAttempts(1))
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

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
	defer tx.Rollback()

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

type Fic struct {
	ID         string
	Created    time.Time
	Edited     time.Time
	AuthorID   string
	Title      string
	ChapterIDs []string

	Description     string `datastore:",noindex"`
	DescriptionHTML string `datastore:",noindex"`
}

func (d *Database) GetFic(id string) (*Fic, error) {
	tx, err := d.ds.NewTransaction(d.ctx, datastore.MaxAttempts(1))
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	fic, _, err := d.getFic(d.ctx, tx, id)
	if err != nil {
		return nil, err
	}

	_, err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return fic, nil
}

func (d *Database) getFic(ctx context.Context, tx *datastore.Transaction, id string) (*Fic, *datastore.Key, error) {
	results := []Fic{}
	var result *Fic
	var resultKey *datastore.Key

	q := datastore.NewQuery("Fic").
		Filter("ID =", id).
		Limit(1).
		Transaction(tx)

	resultKeys, err := d.ds.GetAll(ctx, q, results)
	if err != nil {
		switch err {
		case datastore.ErrNoSuchEntity:
			return nil, nil, ErrNoFicFound
		default:
			return nil, nil, err
		}
	}

	result = &results[0]
	resultKey = resultKeys[0]

	return result, resultKey, nil
}

func (d *Database) putFic(ctx context.Context, tx *datastore.Transaction, fic *Fic) (*datastore.PendingKey, error) {
	_, key, err := d.getFic(ctx, tx, fic.ID)
	if err != nil && err != ErrNoFicFound {
		return nil, err
	}

	if key == nil {
		key = datastore.NewIncompleteKey(ctx, "Fic", nil)
	}

	return tx.Put(key, fic)
}

type Chapter struct {
	ID       string
	FicID    string
	AuthorID string
	Title    string
	Created  time.Time

	Content     string `datastore:",noindex"`
	ContentHTML string `datastore:",noindex"`
}

func (d *Database) addFicChapter(ctx context.Context, tx *datastore.Transaction, fic *Fic, ficKey *datastore.Key, chapter *Chapter) (*datastore.PendingKey, error) {
	fic.ChapterIDs = append(fic.ChapterIDs, chapter.ID)

	_, err := d.putFic(ctx, tx, fic)
	if err != nil {
		return nil, err
	}

	key := datastore.NewIncompleteKey(ctx, "Chapter", ficKey)

	return tx.Put(key, chapter)
}
