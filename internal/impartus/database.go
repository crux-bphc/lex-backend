package impartus

import (
	"log"
	"os"
	"time"

	"github.com/surrealdb/surrealdb.go"
)

type impartusRepository struct {
	DB *surrealdb.DB
}

var Repository *impartusRepository

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

	Repository = &impartusRepository{db}
}

type User struct {
	surrealdb.Basemodel `table:"user"`
	ID                  string    `json:"id,omitempty"`
	EMail               string    `json:"email,omitempty"`
	Jwt                 string    `json:"jwt,omitempty"`
	Password            string    `json:"password,omitempty"`
	UpdatedAt           time.Time `json:"updated_at,omitempty"`
}
