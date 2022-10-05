package users

import (
	"context"
	_ "embed"
	encore "encore.dev"
	"encore.dev/storage/sqldb"
	"log"
	"reflect"
	"testing"
)

//go:embed testdata/fixtures.sql
var fixtures string

func init() {
	if encore.Meta().Environment.Type == encore.EnvLocal {
		if _, err := sqldb.Exec(context.Background(), fixtures); err != nil {
			log.Fatalln("unable to add fixtures:", err)
		}
	}
}

func TestUsersAndFindThemInList(t *testing.T) {
	user, err := Get(context.Background(), 1)
	if err != nil {
		t.Fatal("failed to get user", err)
	}
	expected := Users{Items: []User{*user}}
	actual, err := List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %q to match %q", expected, actual)
	}
}
