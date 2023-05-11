package db

import (
	"context"

	bolt "go.etcd.io/bbolt"
)

// SuperLogger represents logger module (modules/logger)
// you don't need to depend on the particular structures, instead, just declare interface in-place (like Go-way)
type SuperLogger interface {
	SuperLogToStdOut(message string)
}

type DB struct {
	logger SuperLogger
	boltdb *bolt.DB
	path   string
}

type Repository interface {
	Insert()
	Update()
	Delete()
	Select()
}

// Init the plugin
func (db *DB) Init(logger SuperLogger) error {
	logger.SuperLogToStdOut("initializing DB")
	db.logger = logger
	db.path = "./examples_bolt_db"
	bdb, err := bolt.Open(db.path, 0666, nil)
	if err != nil {
		return err
	}

	db.boltdb = bdb
	return nil
}

func (db *DB) Serve() chan error {
	errCh := make(chan error)
	db.logger.SuperLogToStdOut("serving DB")
	return errCh
}

func (db *DB) Stop(context.Context) error {
	return db.boltdb.Close()
}

func (db *DB) Name() string {
	return "super DATABASE service"
}

/////////////// DB LAYER /////////////////

func (db *DB) Insert() {
	db.logger.SuperLogToStdOut("INSERTING")
}

func (db *DB) Update() {
	db.logger.SuperLogToStdOut("UPDATING")
}

func (db *DB) Delete() {
	db.logger.SuperLogToStdOut("DELETING")
}

func (db *DB) Select() {
	db.logger.SuperLogToStdOut("SELECTING")
}
