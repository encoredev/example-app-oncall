package schedules

import (
	"context"
	"encore.app/users"
	"encore.dev/storage/sqldb"
	"errors"
	"fmt"
	"time"
)

type Schedules struct {
	Items []Schedule
}

type Schedule struct {
	Id   int32
	User users.User
	Time TimeRange
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}

//encore:api public method=POST path=/users/:userId/schedules
func Create(ctx context.Context, userId int32, timeRange TimeRange) (*Schedule, error) {
	if timeRange.Start.Before(time.Now()) {
		return nil, fmt.Errorf("start timestamp in the past")
	}

	err := VerifyTimeRange(timeRange)
	if err != nil {
		return nil, err
	}

	user, err := users.Get(ctx, userId)
	if err != nil {
		return nil, err
	}

	// check for existing schedules. we only support 1 at a timestamp right now.
	if s, err := ScheduledAtTime(ctx, timeRange.Start); s != nil && err == nil {
		return nil, fmt.Errorf("schedule already exists within this start timestamp")
	}

	if s, err := ScheduledAtTime(ctx, timeRange.End); s != nil && err == nil {
		return nil, fmt.Errorf("schedule already exists within this end timestamp")
	}

	schedule := Schedule{User: *user, Time: TimeRange{}}
	err = sqldb.QueryRow(
		ctx,
		`INSERT INTO schedules (user_id, start_time, end_time) VALUES ($1, $2, $3) RETURNING id, start_time, end_time`,
		userId, timeRange.Start, timeRange.End,
	).Scan(&schedule.Id, &schedule.Time.Start, &schedule.Time.End)

	if err != nil {
		return nil, err
	}

	return &schedule, nil
}

//encore:api public method=GET path=/scheduled
func ScheduledNow(ctx context.Context) (*Schedule, error) {
	return ScheduledAtTime(ctx, time.Now())
}

//encore:api public method=GET path=/scheduled/:timestamp
func ScheduledAtTimestamp(ctx context.Context, timestamp string) (*Schedule, error) {
	parsedtime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return nil, fmt.Errorf("timestamp is not in a valid format")
	}

	return ScheduledAtTime(ctx, parsedtime)
}

func ScheduledAtTime(ctx context.Context, parsedtime time.Time) (*Schedule, error) {
	schedule, err := RowToSchedule(ctx, sqldb.QueryRow(ctx, `SELECT id, user_id, start_time, end_time FROM schedules WHERE $1 >= start_time AND $1 <= end_time`, parsedtime))
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, fmt.Errorf("no schedule found")
	}
	if err != nil {
		return nil, err
	}
	return schedule, nil
}

//encore:api public method=GET path=/schedules/:id
func GetById(ctx context.Context, id int32) (*Schedule, error) {
	schedule, err := RowToSchedule(ctx, sqldb.QueryRow(ctx, `SELECT id, user_id, start_time, end_time FROM schedules WHERE id = $1`, id))
	if err != nil {
		return nil, fmt.Errorf("schedule not found")
	}

	return schedule, nil
}

//encore:api public method=GET path=/schedules
func ListByTimeRange(ctx context.Context, timeRange TimeRange) (*Schedules, error) {
	var rows *sqldb.Rows
	var err error

	err = VerifyTimeRange(timeRange)
	if err != nil {
		return nil, err
	}

	rows, err = sqldb.Query(
		ctx,
		`SELECT id, user_id, start_time, end_time FROM schedules WHERE start_time > $1 AND end_time < $2 ORDER BY start_time ASC`,
		timeRange.Start, timeRange.End,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var schedules []Schedule
	for rows.Next() {
		schedule, err := RowToSchedule(ctx, rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, *schedule)
	}

	return &Schedules{Items: schedules}, nil
}

//encore:api public method=DELETE path=/schedules/:id
func DeleteById(ctx context.Context, id int32) (*Schedule, error) {
	schedule, err := GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	_, err = sqldb.Exec(ctx, `DELETE FROM schedules WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}

	return schedule, err
}

//encore:api public method=DELETE path=/schedules
func DeleteByTimeRange(ctx context.Context, timeRange TimeRange) (*Schedules, error) {
	schedules, err := ListByTimeRange(ctx, timeRange)
	if err != nil {
		return nil, err
	}

	_, err = sqldb.Exec(ctx, `DELETE FROM schedules WHERE start_time >= $1 AND end_time <= $2`, timeRange.Start, timeRange.End)
	if err != nil {
		return nil, err
	}

	return schedules, err
}

// Helper function from Row to Schedule
func RowToSchedule(ctx context.Context, row interface {
	Scan(dest ...interface{}) error
}) (*Schedule, error) {
	var schedule = &Schedule{Time: TimeRange{}}
	var userId int32

	err := row.Scan(&schedule.Id, &userId, &schedule.Time.Start, &schedule.Time.End)
	if err != nil {
		return nil, err
	}

	user, err := users.Get(ctx, userId)
	if err != nil {
		return nil, err
	}

	schedule.User = *user
	return schedule, nil
}

// Helper function for making sure start and end times are valid
func VerifyTimeRange(timeRange TimeRange) error {
	if timeRange.Start.Equal(timeRange.End) {
		return fmt.Errorf("start timestamp cannot be equal to end timestamp")
	}

	if timeRange.Start.After(timeRange.End) {
		return fmt.Errorf("start timestamp cannot be greater than end timestamp")
	}

	return nil
}
