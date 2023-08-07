package main

import (
	"reflect"
	"testing"
	"time"
)

func Test_getEvent(t *testing.T) {
	type args struct {
		event string
	}
	type wantEvent struct {
		Title string
		Start string
		Stop  string
	}

	tests := []struct {
		name    string
		args    args
		want    wantEvent
		wantErr bool
	}{
		{
			"valid",
			args{
				event: "Day Care pickup\n    Aug 9, 2023 at 16:30 -0700 - 17:00 -0700\n    uid: 2870243A-81F4-4276-A1E3-94F1F5B47139\n",
			},
			wantEvent{
				Title: "Day Care pickup",
				Start: "Aug 9, 2023 16:30 -0700",
				Stop:  "Aug 9, 2023 17:00 -0700",
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

			wantStart, err := time.Parse(timeLayout, tt.want.Start)
			if err != nil {
				t.Fatalf("Invalid Start given: %s, %s", tt.want.Start, err)
			}
			wantStop, err := time.Parse(timeLayout, tt.want.Stop)
			if err != nil {
				t.Fatalf("Invalid Stop given: %s, %s", tt.want.Start, err)
			}

			wantEvent := Event{
				Title: got.Title,
				Start: wantStart,
				Stop:  wantStop,
			}

			if !reflect.DeepEqual(got, wantEvent) {
				t.Errorf("getEvent() = %v, want %v", got, wantEvent)
			}
		})
	}
}
