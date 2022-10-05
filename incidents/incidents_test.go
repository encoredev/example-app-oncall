package incidents

import (
	"context"
	"reflect"
	"testing"
	"time"

	"encore.app/schedules"
	"encore.app/users"
)

func TestCreateIncidents(t *testing.T) {
	user := createUser(t)
	incident := createIncident(t, "Incident #3. This should not be assigned!")

	incidentEquals(t, incident, Incident{
		Body:           "Incident #3. This should not be assigned!",
		Assignee:       nil,
		Acknowledged:   false,
		AcknowledgedAt: nil,
	})

	duration := time.Duration(1000 * 1000)
	createSchedule(t, user, time.Now().Add(duration))

	incident2 := createIncident(t, "Incident #4. This should be assigned to user #2!")
	incidentEquals(t, incident2, Incident{
		Body:           "Incident #4. This should be assigned to user #2!",
		Assignee:       user,
		Acknowledged:   false,
		AcknowledgedAt: nil,
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

func createSchedule(t *testing.T, user *users.User, startTime time.Time) *schedules.Schedule {
	schedule, err := schedules.Create(context.Background(), user.Id, schedules.TimeRange{
		Start: startTime.UTC(),
		End:   startTime.UTC().Add(time.Duration(5000 * 1000 * 1000)),
	})
	if err != nil {
		t.Fatal("failed to create schedule", err)
	}
	return schedule
}

func createIncident(t *testing.T, body string) *Incident {
	incident, err := Create(context.Background(), &CreateParams{Body: body})
	if err != nil {
		t.Fatal(err)
	}
	return incident
}

func incidentEquals(t *testing.T, actual *Incident, expected Incident) {
	if actual.Body != expected.Body {
		t.Errorf("Body does not match provided value. got %q, want %q", actual.Body, expected.Body)
	}

	if !reflect.DeepEqual(actual.Assignee, expected.Assignee) {
		t.Errorf("Assignee does not match provided value. got %q, want %q", actual.Assignee, expected.Assignee)
	}

	if actual.Acknowledged != expected.Acknowledged {
		t.Errorf("Acknowledged does not match provided value. got %q, want %q", actual.Acknowledged, expected.Acknowledged)
	}

	if actual.AcknowledgedAt != expected.AcknowledgedAt {
		t.Errorf("AcknowledgedAt does not match provided value. got %q, want %q", actual.AcknowledgedAt, expected.AcknowledgedAt)
	}
}
