package db

import (
	"github.com/spiral/cascade/samples/db_http_logger/modules/logger"
	bolt "go.etcd.io/bbolt"
)

type DB struct {
	logger logger.SuperLogger
	boltdb *bolt.DB
	path   string
}

type Repository interface {
	Insert()
	Update()
	Delete()
	Select()
}

func (db *DB) Init(logger logger.SuperLogger) error {
	db.logger = logger
	db.path = "./samples/db_http_logger/modules/db/sample"
	bdb, err := bolt.Open(db.path, 0666, nil)
	if err != nil {
		return err
	}

	db.boltdb = bdb
	return nil
}

func (db *DB) Serve() chan error {
	errCh := make(chan error)
	db.logger.SuperLogToStdOut("start serving DB")
	return errCh
}

func (db *DB) Configure() error {
	return nil
}

func (db *DB) Close() error {
	return db.boltdb.Close()
}

func (db *DB) Stop() error {
	return nil
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