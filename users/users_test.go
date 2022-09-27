package users

import (
	"context"
	"reflect"
	"testing"
)

func TestUsersAndFindThemInList(t *testing.T) {
	user := createUser(t)
	expected := Users{Items: []User{*user}}
	actual, err := List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %q to match %q", expected, actual)
	}
}

func createUser(t *testing.T) *User {
	user, err := Create(context.Background(), CreateParams{
		FirstName:   "Bilawal",
		LastName:    "Hameed",
		SlackHandle: "bil",
	})
	if err != nil {
		t.Fatal("failed to create user", err)
	}
	return user
}
