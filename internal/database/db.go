package database

import (
	"log"
	"os"

	"github.com/surrealdb/surrealdb.go"
)

var impartus *impartusRepository

func init() {

	db, err := surrealdb.New("ws://db:8000/rpc")
	if err != nil {
		log.Fatalln(err)
	}

	if _, err = db.Signin(map[string]interface{}{
		"user": os.Getenv("DB_USER"),
		"pass": os.Getenv("DB_PASSWORD"),
	}); err != nil {
		log.Fatalln(err)

	}

	if _, err = db.Use(os.Getenv("DB_NAMESPACE"), "impartus"); err != nil {
		log.Fatalln(err)
	}

	impartus = &impartusRepository{db}
}

func GetImpartusRepository() *impartusRepository {
	return impartus
}
