package users

//func TestUsersAndFindThemInList(t *testing.T) {
//	user := createUser(t)
//	expected := Users{Items: []User{*user}}
//	actual, err := List(context.Background())
//	if err != nil {
//		t.Fatal(err)
//	}
//	if reflect.DeepEqual(actual, expected) {
//		t.Errorf("expected %q to match %q", expected, actual)
//	}
//}
//
//func contains(items []User, expected User) bool {
//	for _, item := range items {
//		if item == expected {
//			return true
//		}
//	}
//	return false
//}
//
//func createUser(t *testing.T) *User {
//	user, err := Create(context.Background(), CreateParams{
//		FirstName:   "Bilawal",
//		LastName:    "Hameed",
//		SlackHandle: "bil",
//	})
//	if err != nil {
//		t.Fatalf("failed to create user %q", err)
//	}
//	return user
//}
