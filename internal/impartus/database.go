package impartus

import (
	"errors"
	"log"
	"os"
	"time"

	surrealdb "github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"
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

	if _, err = db.SignIn(&surrealdb.Auth{
		Username: os.Getenv("SURREAL_USER"),
		Password: os.Getenv("SURREAL_PASS"),
	}); err != nil {
		log.Fatalln(err)

	}

	if err = db.Use("lex", "impartus"); err != nil {
		log.Fatalln(err)
	}

	Repository = &impartusRepository{db}
}

type User struct {
	ID        *models.RecordID `json:"id,omitempty"`
	EMail     string           `json:"email,omitempty"`
	Jwt       string           `json:"jwt,omitempty"`
	Password  string           `json:"password,omitempty"`
	UpdatedAt time.Time        `json:"updated_at,omitempty"`
}

type Subject struct {
	ID         *models.RecordID `json:"id,omitempty"`
	Department string           `json:"department,omitempty"`
	Code       string           `json:"code,omitempty"`
	Name       string           `json:"name,omitempty"`
}

type Lecture struct {
	ID              *models.RecordID `json:"id,omitempty"`
	ImpartusSession int              `json:"impartus_session,omitempty"`
	ImpartusSubject int              `json:"impartus_subject,omitempty"`
	Section         int              `json:"section,omitempty"`
	Professor       string           `json:"professor,omitempty"`
	Users           []string         `json:"-"`
}

// TODO: add ability to search for and filter lectures
func (repo *impartusRepository) GetSubjects(query string) ([]Subject, error) {
	res, err := surrealdb.Query[[]Subject](
		repo.DB,
		`SELECT *, search::score(1) as search_score OMIT search_score 
		FROM subject WHERE name @1@ $query ORDER BY score DESC LIMIT 15`,
		map[string]interface{}{
			"query": query,
		},
	)
	return (*res)[0].Result, err
}

// Get the list of subjects that are pinned by the user
func (repo *impartusRepository) GetPinnedSubjects(email string) ([]Subject, error) {
	res, err := surrealdb.Query[[]Subject](
		repo.DB,
		"SELECT * from $user->pinned->subject",
		map[string]interface{}{
			"user": models.RecordID{Table: "user", ID: email},
		},
	)
	if err != nil {
		return nil, err
	}
	return (*res)[0].Result, nil
}

// Get a valid impartus jwt token of a user who is registered to the lecture
func (repo *impartusRepository) GetLectureToken(sessionId, subjectId int) (string, error) {
	res, err := surrealdb.Query[string](repo.DB,
		"array::first((SELECT VALUE fn::get_token(id) from $lecture.users)[WHERE !type::is::none($this)])",
		map[string]interface{}{
			"lecture": models.RecordID{Table: "lecture", ID: []int{sessionId, subjectId}},
		},
	)

	if err != nil {
		return "", err
	}

	data := (*res)[0].Result

	if len(data) == 0 {
		return "", errors.New("no valid user is registered under this course")
	}

	return data, nil
}
