package incidents

import (
	"context"
	"reflect"
	"testing"
	"time"

	"encore.app/schedules"
	"encore.app/users"
)

func TestIncidents(t *testing.T) {
	t.Run("with no assignee", incidentDeepEquals(Incident{
		Body:           "The first incident. This shouldn't be assigned!",
		Assignee:       nil,
		Acknowledged:   false,
		AcknowledgedAt: nil,
	}))

	t.Run("with assignee", func(t *testing.T) {
		var user = createUser(t)
		var _ = createSchedule(t, user)

		time.Sleep(time.Duration(100 * 1000 * 1000))

		incidentDeepEquals(Incident{
			Body:           "The second incident. This should be assigned!",
			Assignee:       user,
			Acknowledged:   false,
			AcknowledgedAt: nil,
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
		t.Fatalf("failed to create user %q", err)
	}
	return user
}

func createSchedule(t *testing.T, user *users.User) *schedules.Schedule {
	schedule, err := schedules.Create(context.Background(), user.Id, schedules.TimeRange{
		Start: time.Now().Add(time.Duration(100 * 1000 * 1000)),
		End:   time.Now().Add(time.Duration(1000 * 1000 * 1000)),
	})
	if err != nil {
		t.Fatalf("failed to create schedule %v", err)
	}
	return schedule
}

func incidentDeepEquals(expected Incident) func(t *testing.T) {
	return func(t *testing.T) {
		incident, err := Create(context.Background(), &CreateParams{Body: expected.Body})
		if err != nil {
			t.Fatal(err)
		}
		if incident.Body != expected.Body {
			t.Errorf("response incident.Body does not match provided Body. got %q, want %q", incident.Body, expected.Body)
		}

		if !reflect.DeepEqual(incident.Assignee, expected.Assignee) {
			t.Errorf("incident.Assignee does not match. got %q, want %q", incident.Assignee, expected.Assignee)
		}

		if incident.Acknowledged != expected.Acknowledged {
			t.Errorf("incident.Acknowledged does not match. got %q, want %q", incident.Acknowledged, expected.Acknowledged)
		}

		if incident.AcknowledgedAt != expected.AcknowledgedAt {
			t.Errorf("incident.AcknowledgedAt does not match. got %q, want %q", incident.AcknowledgedAt, expected.AcknowledgedAt)
		}
	}
}
