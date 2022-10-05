package users

import (
	"context"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"errors"
)

type Users struct {
	Items []User
}

type User struct {
	Id          int
	FirstName   string
	LastName    string
	SlackHandle string
}

//encore:api public method=POST path=/users
func Create(ctx context.Context, params CreateParams) (*User, error) {
	eb := errs.B().Meta("params", params)

	if len(params.FirstName) == 0 {
		return nil, eb.Code(errs.InvalidArgument).Msg("first name is empty").Err()
	}

	if len(params.LastName) == 0 {
		return nil, eb.Code(errs.InvalidArgument).Msg("last name is empty").Err()
	}

	if len(params.SlackHandle) == 0 {
		return nil, eb.Code(errs.InvalidArgument).Msg("slack handle is empty").Err()
	}

	user := User{}
	err := sqldb.QueryRow(ctx, `
		INSERT INTO users (first_name, last_name, slack_handle)
		VALUES ($1, $2, $3)
		RETURNING id, first_name, last_name, slack_handle
	`, params.FirstName, params.LastName, params.SlackHandle).Scan(&user.Id, &user.FirstName, &user.LastName, &user.SlackHandle)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

type CreateParams struct {
	FirstName   string
	LastName    string
	SlackHandle string
}

//encore:api public method=GET path=/users/:id
func Get(ctx context.Context, id int) (*User, error) {
	eb := errs.B().Meta("userId", id)

	user := User{}
	err := sqldb.QueryRow(ctx, `
		SELECT id, first_name, last_name, slack_handle
		FROM users
		WHERE id = $1
	`, id).Scan(&user.Id, &user.FirstName, &user.LastName, &user.SlackHandle)

	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, eb.Code(errs.InvalidArgument).Msg("no user found").Err()
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

//encore:api public method=GET path=/users
func List(ctx context.Context) (*Users, error) {
	eb := errs.B()
	rows, err := sqldb.Query(ctx, `
		SELECT id, first_name, last_name, slack_handle
		FROM users
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []User
	for rows.Next() {
		var user = User{}
		if err := rows.Scan(&user.Id, &user.FirstName, &user.LastName, &user.SlackHandle); err != nil {
			return nil, eb.Code(errs.Unknown).Msgf("could not scan: %v", err).Err()
		}
		users = append(users, user)
	}

	return &Users{Items: users}, nil
}
