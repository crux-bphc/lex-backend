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
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	}); err != nil {
		log.Fatalln(err)

	}

	if err = db.Use(os.Getenv("DB_NAMESPACE"), "impartus"); err != nil {
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
func (repo *impartusRepository) GetSubjects() ([]Subject, error) {
	res, err := surrealdb.Select[[]Subject](repo.DB, models.Table("subject"))
	return (*res), err
}

// Get the list of subjects that are pinned by the user
func (repo *impartusRepository) GetPinnedSubjects(email string) ([]Subject, error) {
	res, err := surrealdb.Query[[]Subject](
		repo.DB,
		"SELECT * from subject where <-pinned<-(user where email = $email)",
		map[string]interface{}{"email": email},
	)
	if err != nil {
		return nil, err
	}
	return (*res)[0].Result, nil
}

// Get all lecture sections corresponding to a particular subject
func (repo *impartusRepository) GetLectures(deparment, subjectCode string) ([]Lecture, error) {
	res, err := surrealdb.Query[[]Lecture](
		repo.DB,
		"SELECT * FROM lecture WHERE subject = $subject",
		map[string]interface{}{"subject": models.RecordID{Table: "subject", ID: []string{deparment, subjectCode}}},
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
		map[string]interface{}{"lecture": models.RecordID{Table: "lecture", ID: []int{sessionId, subjectId}}},
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
