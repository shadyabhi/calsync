package maccalendar

import (
	"calsync/event"
	"reflect"
	"strings"
	"testing"
	"time"
)

func Test_GetEvent(t *testing.T) {
	type args struct {
		event string
	}
	type wantEvent struct {
		Title string
		Start string
		Stop  string
		UID   string
	}

	tests := []struct {
		name    string
		args    args
		want    wantEvent
		wantErr bool
	}{
		{
			"valid - today's event",
			args{
				event: "Day Care drop\n    today at 08:30 -0700 - 09:05 -0700\n    uid: CC0D3F1F-8703-4783-BF24-526D1DD10618\n",
			},
			wantEvent{
				Title: "Day Care pickup",
				Start: "today 08:30 -0700",
				Stop:  "today 09:05 -0700",
				UID:   "CC0D3F1F-8703-4783-BF24-526D1DD10618",
			},
			false,
		},
		{
			"valid - tomorrow's event",
			args{
				event: "Day Care drop\n    tomorrow at 08:30 -0700 - 09:05 -0700\n    uid: CC0D3F1F-8703-4783-BF24-526D1DD10618\n",
			},
			wantEvent{
				Title: "Day Care pickup",
				Start: "tomorrow 08:30 -0700",
				Stop:  "tomorrow 09:05 -0700",
				UID:   "CC0D3F1F-8703-4783-BF24-526D1DD10618",
			},
			false,
		},
		{
			"valid - day after tomorrow's event",
			args{
				event: "Day Care drop\n    day after tomorrow at 08:30 -0700 - 09:05 -0700\n    uid: CC0D3F1F-8703-4783-BF24-526D1DD10618\n",
			},
			wantEvent{
				Title: "Day Care pickup",
				Start: "day after tomorrow 08:30 -0700",
				Stop:  "day after tomorrow 09:05 -0700",
				UID:   "CC0D3F1F-8703-4783-BF24-526D1DD10618",
			},
			false,
		},
		{
			"valid - any date",
			args{
				event: "Day Care pickup\n    Aug 9, 2023 at 16:30 -0700 - 17:00 -0700\n    uid: 2870243A-81F4-4276-A1E3-94F1F5B47139\n",
			},
			wantEvent{
				Title: "Day Care pickup",
				Start: "Aug 9, 2023 16:30 -0700",
				Stop:  "Aug 9, 2023 17:00 -0700",
				UID:   "2870243A-81F4-4276-A1E3-94F1F5B47139",
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getEvent(tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("getEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Start.IsZero() || got.Stop.IsZero() {
				t.Fatalf("Start/Stop fields in Event can't be zero: %#v", got)
			}

			// Handle common name for date in tests
			now := time.Now().Local()
			tt.want.Start = strings.Replace(tt.want.Start, "today", now.Format("Jan 2, 2006"), 1)
			tt.want.Stop = strings.Replace(tt.want.Stop, "today", now.Format("Jan 2, 2006"), 1)

			if !strings.Contains(tt.want.Start, "day after") {
				tt.want.Start = strings.Replace(tt.want.Start, "tomorrow", now.Add(24*time.Hour).Format("Jan 2, 2006"), 1)
			} else {
				tt.want.Start = strings.Replace(tt.want.Start, "day after tomorrow", now.Add(48*time.Hour).Format("Jan 2, 2006"), 1)

			}
			if !strings.Contains(tt.want.Stop, "day after") {
				tt.want.Stop = strings.Replace(tt.want.Stop, "tomorrow", now.Add(24*time.Hour).Format("Jan 2, 2006"), 1)
			} else {
				tt.want.Stop = strings.Replace(tt.want.Stop, "day after tomorrow", now.Add(48*time.Hour).Format("Jan 2, 2006"), 1)
			}

			wantStart, err := time.Parse(timeLayout, tt.want.Start)
			if err != nil {
				t.Fatalf("Invalid Start given: %s, %s", tt.want.Start, err)
			}
			wantStop, err := time.Parse(timeLayout, tt.want.Stop)
			if err != nil {
				t.Fatalf("Invalid Stop given: %s, %s", tt.want.Start, err)
			}

			wantEvent := event.Event{
				Title: got.Title,
				Start: wantStart,
				Stop:  wantStop,
				UID:   tt.want.UID,
			}

			if !reflect.DeepEqual(got, wantEvent) {
				t.Errorf("getEvent() = %v, want %v", got, wantEvent)
			}
		})
	}
}
