package schedules

//func TestSchedules(t *testing.T) {
//	var timeRange *TimeRange
//	var user *users.User
//	var schedule *Schedule
//
//	var wideTimeRange = TimeRange{
//		Start: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
//		End:   time.Date(2099, 12, 31, 23, 59, 0, 0, time.UTC),
//	}
//
//	t.Run("with an empty list", func(t *testing.T) {
//		if schedules, _ := ListByTimeRange(context.Background(), wideTimeRange); schedules.Items != nil {
//			t.Fatal("schedule should be an empty list")
//		}
//	})
//
//	t.Run("with nobody scheduled for now", func(t *testing.T) {
//		timeRange = &TimeRange{
//			Start: time.Now().Add(time.Duration(100 * 1000 * 1000)),
//			End:   time.Now().Add(time.Duration(1000 * 1000 * 1000)),
//		}
//		user = createUser(t)
//		schedule = createSchedule(t, user, *timeRange)
//
//		if got, _ := ScheduledAtTime(context.Background(), time.Now()); got != nil {
//			t.Errorf("expecting schedule to be currently empty")
//		}
//	})
//
//	t.Run("with someone scheduled for now", func(t *testing.T) {
//		time.Sleep(time.Duration(100 * 1000 * 1000))
//
//		if got, _ := ScheduledAtTime(context.Background(), timeRange.Start); !reflect.DeepEqual(got, schedule) {
//			t.Errorf("got %q, want %q", got, schedule)
//		}
//	})
//
//	t.Run("with one item in the list", func(t *testing.T) {
//		if schedules, _ := ListByTimeRange(context.Background(), wideTimeRange); len(schedules.Items) != 1 {
//			t.Fatal("schedules should be an empty list")
//		}
//	})
//}
//
//func createUser(t *testing.T) *users.User {
//	user, err := users.Create(context.Background(), users.CreateParams{
//		FirstName:   "Bilawal",
//		LastName:    "Hameed",
//		SlackHandle: "bil",
//	})
//	if err != nil {
//		t.Fatalf("failed to create user %q", err)
//	}
//	return user
//}
//
//func createSchedule(t *testing.T, user *users.User, timeRange TimeRange) *Schedule {
//	schedule, err := Create(context.Background(), user.Id, timeRange)
//	if err != nil {
//		t.Fatalf("failed to create schedule %q", err)
//	}
//	return schedule
//}
