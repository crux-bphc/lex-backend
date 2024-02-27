package database_test

import (
	"os"
	"testing"
	"time"

	"github.com/crux-bphc/lex/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/surrealdb/surrealdb.go"
)

// Verify if the lectures are being extracted and added to the database when a new user entry is created
func TestLectureExtraction(t *testing.T) {
	assert := assert.New(t)
	repo := database.GetImpartusRepository()
	_, err := repo.DB.Create("user:populate", database.ImpartusUser{
		EMail:     "f20220149@mudit.com",
		Password:  "za_warudo",
		Jwt:       os.Getenv("IMPARTUS_TEST_TOKEN"),
		UpdatedAt: time.Now(),
	})
	assert.Nil(err)

	subjectCount, err := surrealdb.SmartUnmarshal[int](repo.DB.Query("(select count() from only subject group all limit 1).count", nil))
	assert.Nil(err)
	assert.Greater(subjectCount, 1, "No. of subjects > 1")

	lectureCount, err := surrealdb.SmartUnmarshal[int](repo.DB.Query("(select count() from only lecture where users contains user:populate group all limit 1).count", nil))
	assert.Nil(err)
	assert.Greater(lectureCount, 1, "No. of lectures > 1")

	assert.GreaterOrEqual(lectureCount, subjectCount, "No. of lectures >= subjects")

	pinCount, err := surrealdb.SmartUnmarshal[int](repo.DB.Query("(select count() from only pinned where in = user:populate group all limit 1).count", nil))
	assert.Nil(err)
	assert.Greater(pinCount, 1, "No. of pinned subjects > 1")

	assert.LessOrEqual(pinCount, subjectCount, "No. of pinned subjects is <= the subjects")
}

// Verify if the token is revalidated if it has not been updated since 7 days
func TestTokenRevalidation(t *testing.T) {
	assert := assert.New(t)
	repo := database.GetImpartusRepository()
	_, err := repo.DB.Create("user:revalidate_token", database.ImpartusUser{
		EMail:     "kira_does_dev@crux.com",
		Password:  "WRONG_PASSWORD",
		Jwt:       os.Getenv("IMPARTUS_TEST_TOKEN"),
		UpdatedAt: time.Now(),
	})
	assert.Nil(err)

	oldJwt, err := surrealdb.SmartUnmarshal[string](repo.DB.Query("fn::get_token(user:revalidate_token)", nil))
	assert.Nil(err)
	assert.Equal(os.Getenv("IMPARTUS_TEST_TOKEN"), oldJwt, "The JWT does not update")

	repo.DB.Query("UPDATE user:revalidate_token set updated_at = time::now() - 7d", nil)

	newJwt, err := surrealdb.SmartUnmarshal[string](repo.DB.Query("fn::get_token(user:revalidate_token)", nil))
	assert.Nil(err)

	assert.Empty(newJwt, "The updated JWT should be empty")
}
