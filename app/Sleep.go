package main

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

type Sleep struct {
	ID          int    `json:"id"`
	Date        string `json:"date"`
	Wake        int    `json:"wake"`
	Bath        int    `json:"bath"`
	Bed         int    `json:"bed"`
	Sleep_in    string `json:"sleep_in"`
	Sleep       string `json:"sleep"`
	Deep_sleep  string `json:"deep_sleep"`
	Description string `json:"description"`
}

func getSleep(month string) ([]Sleep, error) {
	var sleeps []Sleep
	db, err := dbConnect()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer func() {
		dbCloseErr := db.Close()
		if dbCloseErr != nil {
			log.Fatal("DB Close Error:", err)
			return
		}
		err = dbCloseErr
	}()

	// yyyymmの形で入ってきます。
	monthScope := getNowMonth(month)
	startDay := getStartDay(monthScope)
	endDay := getEndDay(monthScope)
	//log.Println("startDay : ")
	//log.Println(startDay)
	//log.Println("endDay :")
	//log.Println(endDay)
	query := `
	  SELECT
	    id,
	    date,
	    wake,
	    bath,
	    bed,
	    sleep_in,
	    sleep,
	    deep_sleep,
	    description
	  FROM
	    sleeps
	  WHERE
	    date>=? AND date<=?`
	sleepRows, err := db.Query(query, startDay, endDay)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("query error: %v", err)
	}
	// //ID          int    `json:"id"`
	// //Date        string `json:"date"`
	// //Wake        string `json:"wake"`
	// //Bath        string `json:"bath"`
	// //Bed         string `json:"bed"`
	// //Sleep_in    string `json:"sleep_in"`
	// //Sleep       string `json:"sleep"`
	// //Deep_sleep  string `json:"deep_sleep"`
	// //Description string `json:"description"`
	for sleepRows.Next() {
		var sleep Sleep
		if err := sleepRows.Scan(
			&sleep.ID,
			&sleep.Date,
			&sleep.Wake,
			&sleep.Bath,
			&sleep.Bed,
			&sleep.Sleep_in,
			&sleep.Sleep,
			&sleep.Deep_sleep,
			&sleep.Description); err != nil {
			log.Println(err)
			return nil, fmt.Errorf("scan the sale error: %v", err)
		}
		sleeps = append(sleeps, sleep)
	}
	if err := sleepRows.Err(); err != nil {
		log.Println(err)
		return nil, fmt.Errorf("scan home sale error: %v", err)
	}
	return sleeps, err
}
func getNowMonth(month string) (monthScope string) {
	var ret string
	now := time.Now()
	if month == "" {
		ret = now.Format("2006-01-02")
	} else {
		// yyyymmの形決め打ちで作成する。
		tmpYearStr := month[:4]
		tmpMonthStr := month[4:]
		tmpYearNum, err := strconv.Atoi(tmpYearStr)
		if err != nil {
			log.Fatal("Year Conv Error: ", err)
			log.Fatal("Error Year Str: ", tmpYearStr)
			return
		}
		// TODO
		tmpMonthNum, err := strconv.Atoi(tmpMonthStr)
		if err != nil {
			log.Fatal("Month Conv Error: ", err)
			log.Fatal("Error Month Str: ", tmpMonthStr)
			return
		}
		//log.Println(tmpYearStr)
		//log.Println(tmpMonthStr)
		//log.Println(tmpYearNum)
		//log.Println(tmpMonthNum)
		ret = time.Date(tmpYearNum, time.Month(tmpMonthNum), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	}
	log.Println(ret)
	return ret
}
func getStartDay(month string) (startDay time.Time) {
	now := time.Now()
	tmpMonth := month + " 00:00:00"
	monthDay, err := time.Parse("2006-01-02 15:04:05", tmpMonth)
	if err != nil {
		log.Fatal("first Day Parse Error:", err)
		log.Fatal("Fatal Date:", tmpMonth)
	}
	firstDay := time.Date(monthDay.Year(), monthDay.Month(), 1, 0, 0, 0, 0, now.Location())
	return firstDay
}
func getEndDay(month string) (startDay time.Time) {
	now := time.Now()
	tmpMonth := month + " 23:59:59"
	monthDay, err := time.Parse("2006-01-02 15:04:05", tmpMonth)
	if err != nil {
		log.Fatal("last Day Parse Error:", err)
		log.Fatal("Fatal Date:", tmpMonth)
	}
	lastDay := time.Date(monthDay.Year(), monthDay.Month(), 1, 23, 59, 59, 0, now.Location()).AddDate(0, 1, -1)
	return lastDay
}
