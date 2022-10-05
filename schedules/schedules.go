package schedules

import (
	"context"
	"errors"
	"time"

	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
)

type Schedules struct {
	Items []Schedule
}

type Schedule struct {
	Id     int
	UserId int
	Time   TimeRange
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}

//encore:api public method=POST path=/users/:userId/schedules
func Create(ctx context.Context, userId int, timeRange TimeRange) (*Schedule, error) {
	eb := errs.B().Meta("userID", userId, "start", timeRange.Start.String(), "end", timeRange.End.String())
	if timeRange.Start.Before(time.Now()) {
		return nil, eb.Code(errs.InvalidArgument).Msg("start timestamp in the past").Err()
	}

	err := VerifyTimeRange(timeRange)
	if err != nil {
		return nil, eb.Code(errs.InvalidArgument).Cause(err).Msg("invalid time range").Err()
	}

	// check for existing schedules. we only support 1 at a timestamp right now.
	if schedule, err := ScheduledAt(ctx, timeRange.Start.Format(time.RFC3339)); schedule != nil && err == nil {
		return nil, eb.Code(errs.InvalidArgument).Cause(err).Msg("schedule already exists within this start timestamp").Err()
	}

	if schedule, err := ScheduledAt(ctx, timeRange.End.Format(time.RFC3339)); schedule != nil && err == nil {
		return nil, eb.Code(errs.InvalidArgument).Cause(err).Msg("schedule already exists within this end timestamp").Err()
	}

	schedule := Schedule{UserId: userId, Time: TimeRange{}}
	err = sqldb.QueryRow(ctx, `
		INSERT INTO schedules (user_id, start_time, end_time)
		VALUES ($1, $2, $3)
		RETURNING id, start_time, end_time
	`, userId, timeRange.Start, timeRange.End).Scan(&schedule.Id, &schedule.Time.Start, &schedule.Time.End)
	if err != nil {
		return nil, eb.Code(errs.Unavailable).Cause(err).Msg("insert schedule").Err()
	}

	return &schedule, nil
}

//encore:api public method=GET path=/scheduled
func ScheduledNow(ctx context.Context) (*Schedule, error) {
	return Scheduled(ctx, time.Now())
}

//encore:api public method=GET path=/scheduled/:timestamp
func ScheduledAt(ctx context.Context, timestamp string) (*Schedule, error) {
	eb := errs.B().Meta("timestamp", timestamp)
	parsedtime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return nil, eb.Code(errs.InvalidArgument).Msg("timestamp is not in a valid format").Err()
	}

	return Scheduled(ctx, parsedtime)
}

func Scheduled(ctx context.Context, timestamp time.Time) (*Schedule, error) {
	eb := errs.B().Meta("timestamp", timestamp.String())
	schedule, err := RowToSchedule(ctx, sqldb.QueryRow(ctx, `
		SELECT id, user_id, start_time, end_time
		FROM schedules
		WHERE start_time <= $1
		  AND end_time >= $1
	`, timestamp.UTC()))
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, eb.Code(errs.NotFound).Msg("no schedule found").Err()
	}
	if err != nil {
		return nil, err
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

	rows, err = sqldb.Query(ctx, `
		SELECT id, user_id, start_time, end_time
		FROM schedules
		WHERE start_time >= $1
		  AND end_time <= $2
		ORDER BY start_time ASC
	`, timeRange.Start, timeRange.End)
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

// RowToSchedule Helper function from Row to Schedule
func RowToSchedule(ctx context.Context, row interface {
	Scan(dest ...interface{}) error
}) (*Schedule, error) {
	var schedule = &Schedule{Time: TimeRange{}}
	err := row.Scan(&schedule.Id, &schedule.UserId, &schedule.Time.Start, &schedule.Time.End)
	if err != nil {
		return nil, err
	}
	return schedule, nil
}

// VerifyTimeRange Helper function for making sure start and end times are valid
func VerifyTimeRange(timeRange TimeRange) error {
	eb := errs.B().Meta("start", timeRange.Start.String(), "end", timeRange.End.String())
	if timeRange.Start.Equal(timeRange.End) {
		return eb.Code(errs.InvalidArgument).Msg("start timestamp cannot be equal to end timestamp").Err()
	}

	if timeRange.Start.After(timeRange.End) {
		return eb.Code(errs.InvalidArgument).Msg("start timestamp cannot be greater than end timestamp").Err()
	}

	return nil
}
