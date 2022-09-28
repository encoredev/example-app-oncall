package incidents

import (
	"context"
	"encore.app/schedules"
	"encore.app/slack"
	"encore.app/users"
	"encore.dev/beta/errs"
	"encore.dev/cron"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"fmt"
	"strings"
	"time"
)

type Incidents struct {
	Items []Incident
}

type Incident struct {
	Id             int32
	Body           string
	CreatedAt      time.Time
	Acknowledged   bool
	AcknowledgedAt *time.Time
	Assignee       *users.User
}

//encore:api public method=GET path=/incidents
func List(ctx context.Context) (*Incidents, error) {
	rows, err := sqldb.Query(ctx, `
		SELECT id, assigned_user_id, body, created_at, acknowledged_at
		FROM incidents
		WHERE acknowledged_at IS NULL
	`)
	if err != nil {
		return nil, err
	}
	return RowsToIncidents(ctx, rows)
}

//encore:api public method=PUT path=/incidents/:id/assign
func Assign(ctx context.Context, id int32, params *AssignParams) (*Incident, error) {
	eb := errs.B().Meta("params", params)
	rows, err := sqldb.Query(ctx, `
		UPDATE incidents
		SET assigned_user_id = $1
		WHERE acknowledged_at IS NULL
		  AND id = $2
		RETURNING id, assigned_user_id, body, created_at, acknowledged_at
	`, params.UserId, id)
	if err != nil {
		return nil, err
	}

	incidents, err := RowsToIncidents(ctx, rows)
	if err != nil {
		return nil, err
	}
	if incidents.Items == nil {
		return nil, eb.Code(errs.NotFound).Msg("no incident found").Err()
	}

	incident := &incidents.Items[0]
	_ = slack.Notify(ctx, &slack.NotifyParams{
		Text: fmt.Sprintf("Incident #%d is re-assigned to %s %s <@%s>\n%s", incident.Id, incident.Assignee.FirstName, incident.Assignee.LastName, incident.Assignee.SlackHandle, incident.Body),
	})

	return incident, err
}

type AssignParams struct {
	UserId int32
}

//encore:api public method=PUT path=/incidents/:id/acknowledge
func Acknowledge(ctx context.Context, id int32) (*Incident, error) {
	eb := errs.B().Meta("incidentId", id)
	rows, err := sqldb.Query(ctx, `
		UPDATE incidents
		SET acknowledged_at = NOW()
		WHERE acknowledged_at IS NULL
		  AND id = $1
		RETURNING id, assigned_user_id, body, created_at, acknowledged_at
	`, id)
	if err != nil {
		return nil, err
	}

	incidents, err := RowsToIncidents(ctx, rows)
	if err != nil {
		return nil, err
	}
	if incidents.Items == nil {
		return nil, eb.Code(errs.NotFound).Msg("no incident found").Err()
	}

	incident := &incidents.Items[0]
	_ = slack.Notify(ctx, &slack.NotifyParams{
		Text: fmt.Sprintf("Incident #%d assigned to %s %s <@%s> has been acknowledged:\n%s", incident.Id, incident.Assignee.FirstName, incident.Assignee.LastName, incident.Assignee.SlackHandle, incident.Body),
	})

	return incident, err
}

//encore:api public method=POST path=/incidents/acknowledge_all
func AcknowledgeAll(ctx context.Context) (*Incident, error) {
	eb := errs.B()
	rows, err := sqldb.Query(ctx, `
		UPDATE incidents
		SET acknowledged_at = NOW()
		WHERE acknowledged_at IS NULL
		RETURNING id, assigned_user_id, body, created_at, acknowledged_at
	`)
	if err != nil {
		return nil, err
	}

	incidents, err := RowsToIncidents(ctx, rows)
	if err != nil {
		return nil, err
	}
	if incidents.Items == nil {
		return nil, eb.Code(errs.NotFound).Msg("no incident found").Err()
	}

	return &incidents.Items[0], err
}

//encore:api public method=POST path=/incidents
func Create(ctx context.Context, params *CreateParams) (*Incident, error) {
	// check who is on-call
	schedule, err := schedules.ScheduledNow(ctx)

	incident := Incident{}
	if schedule != nil {
		incident.Assignee = &schedule.User
	}

	var row *sqldb.Row
	if schedule != nil {
		// Someone is on-call
		row = sqldb.QueryRow(ctx, `
			INSERT INTO incidents (assigned_user_id, body)
			VALUES ($1, $2)
			RETURNING id, body, created_at
		`, &schedule.User.Id, params.Body)
	} else {
		// Nobody is on-call
		row = sqldb.QueryRow(ctx, `
			INSERT INTO incidents (body)
			VALUES ($1)
			RETURNING id, body, created_at
		`, params.Body)
	}

	if err = row.Scan(&incident.Id, &incident.Body, &incident.CreatedAt); err != nil {
		return nil, err
	}

	var text string
	if incident.Assignee != nil {
		text = fmt.Sprintf("Incident #%d created and assigned to %s %s <@%s>\n%s", incident.Id, incident.Assignee.FirstName, incident.Assignee.LastName, incident.Assignee.SlackHandle, incident.Body)
	} else {
		text = fmt.Sprintf("Incident #%d created and unassigned\n%s", incident.Id, incident.Body)
	}
	_ = slack.Notify(ctx, &slack.NotifyParams{Text: text})

	return &incident, nil
}

type CreateParams struct {
	Body string
}

// Helper to take a sqldb.Rows instance and convert it into a list of Incidents
func RowsToIncidents(ctx context.Context, rows *sqldb.Rows) (*Incidents, error) {
	eb := errs.B()

	defer rows.Close()

	var incidents []Incident
	for rows.Next() {
		var incident = Incident{}
		var assignedUserId *int32
		if err := rows.Scan(&incident.Id, &assignedUserId, &incident.Body, &incident.CreatedAt, &incident.AcknowledgedAt); err != nil {
			return nil, eb.Code(errs.Unknown).Msgf("could not scan: %v", err).Err()
		}
		if assignedUserId != nil {
			user, err := users.Get(ctx, *assignedUserId)
			if err != nil {
				return nil, eb.Code(errs.NotFound).Msgf("could not retrieve user for incident %v", assignedUserId).Err()
			}
			incident.Assignee = user
		}
		incident.Acknowledged = incident.AcknowledgedAt != nil
		incidents = append(incidents, incident)
	}

	return &Incidents{Items: incidents}, nil
}

var _ = cron.NewJob("unacknowledged-incidents-reminder", cron.JobConfig{
	Title:    "Notify on Slack about incidents which are not acknowledged",
	Every:    10 * cron.Minute,
	Endpoint: RemindUnacknowledgedIncidents,
})

//encore:api private
func RemindUnacknowledgedIncidents(ctx context.Context) error {
	incidents, err := List(ctx) // we never query for acknowledged incidents
	if err != nil {
		return err
	}
	if incidents == nil {
		return nil
	}

	var items = []string{"These incidents have not been acknowledged yet. Please acknowledge them otherwise you will be reminded every 10 minutes:"}
	for _, incident := range incidents.Items {
		var assignee string
		if incident.Assignee != nil {
			assignee = fmt.Sprintf("%s %s (<@%s>)", incident.Assignee.FirstName, incident.Assignee.LastName, incident.Assignee.SlackHandle)
		} else {
			assignee = "Unassigned"
		}

		items = append(items, fmt.Sprintf("[%s] [#%d] %s", assignee, incident.Id, incident.Body))
	}

	if len(incidents.Items) > 0 {
		_ = slack.Notify(ctx, &slack.NotifyParams{Text: strings.Join(items, "\n")})
	}

	return nil
}

var _ = cron.NewJob("assign-unassigned-incidents", cron.JobConfig{
	Title:    "Assign unassigned incidents to user currently on-call",
	Every:    cron.Minute,
	Endpoint: AssignUnassignedIncidents,
})

//encore:api private
func AssignUnassignedIncidents(ctx context.Context) error {
	// If this code fail, it is either a server error or someone isn't on-call
	schedule, err := schedules.ScheduledNow(ctx)
	if err != nil {
		return err
	}

	incidents, err := List(ctx) // we never query for acknowledged incidents
	if err != nil {
		return err
	}

	for _, incident := range incidents.Items {
		if incident.Assignee != nil {
			continue // this incident has already been assigned
		}

		_, err := Assign(ctx, incident.Id, &AssignParams{UserId: schedule.User.Id})
		if err == nil {
			rlog.Info("OK assigned unassigned incident", "incident", incident, "user", schedule.User)
		} else {
			rlog.Error("FAIL to assign unassigned incident", "incident", incident, "user", schedule.User, "err", err)
			return err
		}
	}

	return nil
}
