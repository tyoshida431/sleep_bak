package main

import (
	"log"
	"testing"
	"time"
)

// func getEndDay(month string) (startDay time.Time, err error) {
func TestGetEndDay(t *testing.T) {
	// okテスト
	//got, err := getEndDay("2026-09-14")

	tmpMonth := "2026-09-30 23:59:59"
	now := time.Now()
	monthDay, err := time.Parse("2006-01-02 15:04:05", tmpMonth)
	if err != nil {
		log.Println("正データ作成失敗: ", err)
	}
	want := time.Date(monthDay.Year(), monthDay.Month(), 1, 23, 59, 59, 0, now.Location()).AddDate(0, 1, -1)

	tests := []struct {
		name       string
		datestring string
		wantErr    bool
	}{
		{"ok", "2026-09-14", false},
		{"bad 年だけ", "2026", true},
		{"bad 月まで", "2026-09", true},
		{"bad 形式違い", "20260914", true},
		{"bad 存在しない日付", "2026-09-31", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := getEndDay(test.datestring)
			if (err != nil) != test.wantErr {
				log.Println(err)
			}
			if test.wantErr {
				return
			}
			if got != want {
				t.Errorf("getEndDay=%s; want %s", got, want)
			}
		})
	}
}
