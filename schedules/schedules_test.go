package schedules

import (
	"context"
	"reflect"
	"testing"
	"time"

	"encore.app/users"
)

func TestSchedules(t *testing.T) {
	var timeRange *TimeRange
	var user *users.User
	var schedule *Schedule

	var wideTimeRange = TimeRange{
		Start: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2099, 12, 31, 23, 59, 0, 0, time.UTC),
	}

	t.Run("group", func(t *testing.T) {
		t.Run("empty the schedule", func(t *testing.T) {
			if _, err := DeleteByTimeRange(context.Background(), wideTimeRange); err != nil {
				t.Fatal("could not empty the schedule", err)
			}
		})

		t.Run("with an empty list", func(t *testing.T) {
			schedules, err := ListByTimeRange(context.Background(), wideTimeRange)
			if err != nil {
				t.Fatal(err)
			}
			if schedules.Items != nil {
				t.Fatal("schedule should be an empty list")
			}
		})

		t.Run("with nobody scheduled for now", func(t *testing.T) {
			timeRange = &TimeRange{
				Start: time.Now().UTC().Add(time.Duration(100 * 1000 * 1000)),
				End:   time.Now().UTC().Add(time.Duration(1000 * 1000 * 1000)),
			}
			user = createUser(t)
			schedule = createSchedule(t, user, *timeRange)

			if got, _ := Scheduled(context.Background(), time.Now()); got != nil {
				t.Fatal("expecting schedule to be currently empty")
			}
		})

		t.Run("with someone scheduled for now", func(t *testing.T) {
			if got, _ := Scheduled(context.Background(), timeRange.Start); !reflect.DeepEqual(got, schedule) {
				t.Fatalf("got %q, want %q", got, schedule)
			}
		})

		t.Run("with one item in the list", func(t *testing.T) {
			if schedules, _ := ListByTimeRange(context.Background(), wideTimeRange); len(schedules.Items) != 1 {
				t.Fatalf("got %q, want %q", len(schedules.Items), 1)
			}
		})
	})
}

func createUser(t *testing.T) *users.User {
	user, err := users.Create(context.Background(), users.CreateParams{
		FirstName:   "Bilawal",
		LastName:    "Hameed",
		SlackHandle: "bil",
	})
	if err != nil {
		t.Fatal("failed to create user", err)
	}
	return user
}

func createSchedule(t *testing.T, user *users.User, timeRange TimeRange) *Schedule {
	schedule, err := Create(context.Background(), user.Id, timeRange)
	if err != nil {
		t.Fatal("failed to create schedule", err)
	}
	return schedule
}
