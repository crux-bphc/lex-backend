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

type Subject struct {
	surrealdb.Basemodel `table:"subject"`
	ID                  string `json:"id,omitempty"`
	Name                string `json:"name,omitempty"`
}

type Lecture struct {
	surrealdb.Basemodel `table:"lecture"`
	ID                  string `json:"id,omitempty"`
	Section             int    `json:"section,omitempty"`
	Professor           string `json:"professor,omitempty"`
}

// TODO: add ability to search for and filter lectures
func (repo *impartusRepository) GetSubjects() ([]Subject, error) {
	data, err := surrealdb.SmartUnmarshal[[]Subject](repo.DB.Select("subject"))
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Get the list of subjects that are pinned by the user
func (repo *impartusRepository) GetPinnedSubjects(email string) ([]Subject, error) {
	data, err := surrealdb.SmartUnmarshal[[]Subject](
		repo.DB.Query(
			"SELECT * from subject where <-pinned<-(user where email = $email)",
			map[string]interface{}{
				"email": email,
			},
		),
	)

	if err != nil {
		return nil, err
	}

	return data, nil
}

// Get all lecture sections corresponding to a particular subject
func (repo *impartusRepository) GetLectures(subjectId string) ([]Lecture, error) {
	data, err := surrealdb.SmartUnmarshal[[]Lecture](
		repo.DB.Query("SELECT * FROM lecture WHERE subject = $subject", map[string]interface{}{
			"subject": subjectId,
		}),
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}
