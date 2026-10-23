package utils

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestDatabaseErrorMappingDoesNotExposeDetail(t *testing.T) {
	tests := []struct {
		code   string
		status int
	}{{"23505", 409}, {"23503", 409}, {"23001", 409}, {"23514", 400}, {"XX000", 500}}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		ResponseError(c, DatabaseError(&pgconn.PgError{Code: tt.code, Message: "private SQL error"}))
		if w.Code != tt.status || strings.Contains(w.Body.String(), "private SQL") {
			t.Fatalf("%s: %d %s", tt.code, w.Code, w.Body)
		}
	}
}
func TestWrappedAppErrorPreservesCode(t *testing.T) {
	err := WrapError("Duplicate email", ErrCodeConflict, errors.New("secret"))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ResponseError(c, DatabaseError(err))
	if w.Code != 409 {
		t.Fatal(w.Code)
	}
}
