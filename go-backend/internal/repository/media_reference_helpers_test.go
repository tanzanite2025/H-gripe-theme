package repository

import (
	"reflect"
	"testing"
)

func TestMediaReferenceTextExpressionUsesDialectCompatibleCasts(t *testing.T) {
	tests := []struct {
		name    string
		dialect string
		want    string
	}{
		{name: "postgres", dialect: "postgres", want: "CAST(images AS TEXT)"},
		{name: "mysql", dialect: "mysql", want: "CAST(images AS CHAR)"},
		{name: "sqlite", dialect: "sqlite", want: "CAST(images AS TEXT)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mediaReferenceTextExpression(tt.dialect, "images"); got != tt.want {
				t.Fatalf("mediaReferenceTextExpression(%q) = %q, want %q", tt.dialect, got, tt.want)
			}
		})
	}
}

func TestMediaReferenceContainsConditionCastsJSONColumnsBeforeLike(t *testing.T) {
	repository := &MediaRepository{}
	query, args := repository.mediaReferenceContainsCondition([]string{"images"}, []string{"/uploads/logo.png"})

	if want := "CAST(images AS TEXT) LIKE ?"; query != want {
		t.Fatalf("media reference query = %q, want %q", query, want)
	}
	if want := []interface{}{"%/uploads/logo.png%"}; !reflect.DeepEqual(args, want) {
		t.Fatalf("media reference args = %#v, want %#v", args, want)
	}
}
