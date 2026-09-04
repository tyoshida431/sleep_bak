package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Sleep struct {
	ID          int    `json:"id"`
	Date        string `json:"date"`
	DateStr     string `json:"date_str"`
	Wake        int    `json:"wake"`
	Bath        int    `json:"bath"`
	Bed         int    `json:"bed"`
	Sleep_in    string `json:"sleep_in"`
	Sleep       string `json:"sleep"`
	Deep_sleep  string `json:"deep_sleep"`
	Description string `json:"description"`
}

type SleepFromFront struct {
	Date        string `json:"date"`
	Wake        string `json:"wake"`
	Bath        string `json:"bath"`
	Bed         string `json:"bed"`
	Sleep_in    string `json:"sleep_in"`
	Sleep       string `json:"sleep"`
	Deep_sleep  string `json:"deep_sleep"`
	Description string `json:"description"`
}

func getSleep(monthFromURLQuery string) ([]Sleep, error) {
	var sleeps []Sleep
	db, err := dbConnect()
	if err != nil {
		log.Println("db Connect Error: ", err)
		return nil, err
	}
	defer func() {
		dbCloseErr := db.Close()
		if dbCloseErr != nil {
			log.Println("DB Close Error:", err)
			return
		}
		err = dbCloseErr
	}()

	// yyyymmの形で入って来るのを決め打ちします。
	month, err := shapeMonth(monthFromURLQuery)
	if err != nil {
		log.Println("shapeMonth error. Invalid Month: ", err)
		return nil, err
	}
	startDay, err := getStartDay(month)
	if err != nil {
		log.Println("Can't get StartDay: ", err)
		return nil, err
	}
	endDay := getEndDay(month)
	if err != nil {
		log.Println("Can't get EndDay: ", err)
		return nil, err
	}
	err = makeNewMonth(db, startDay, endDay)
	if err != nil {
		log.Println("Make New Month Error:", err)
		return nil, err
	}
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
		log.Println("select sleeps query error: ", err)
		return nil, fmt.Errorf("select sleeps query error: %v", err)
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
			log.Println("Sleep Row Scan Error: ", err)
			return nil, fmt.Errorf("scan the sleep error: %v", err)
		}
		sleep.DateStr = changeDateString(sleep.Date)
		sleeps = append(sleeps, sleep)
	}
	if err := sleepRows.Err(); err != nil {
		log.Println("sleep Row Error: ", err)
		return nil, fmt.Errorf("scan sleep error: %v", err)
	}
	return sleeps, err
}
func makeNewMonth(db *sqlx.DB, startDay time.Time, endDay time.Time) error {
	countQuery := `
		SELECT
			COUNT(*)
		FROM
			sleeps
	    WHERE
	    	date>=? AND date<=?`
	countRows, err := db.Query(countQuery, startDay, endDay)
	if err != nil {
		log.Println("sleeps month exist count query error: ", err)
		return fmt.Errorf("sleeps month exist count query error: %v", err)
	}
	var count int
	for countRows.Next() {
		if err := countRows.Scan(&count); err != nil {
			log.Println("sleeps month exist count scan error: ", err)
			return fmt.Errorf("sleeps month exist count scan error: %v", err)
		}
	}
	if err := countRows.Err(); err != nil {
		log.Println("sleeps month exist count rows error: ", err)
		return fmt.Errorf("sleeps month exist count rows error: %v", err)
	}
	if count == 0 {
		year := startDay.Year()
		monthNum := int(startDay.Month())
		dayNum := startDay.Day()
		endDayNum := endDay.Day()
		now := time.Now()
		insertQuery := `
			INSERT INTO sleeps(
				date,
				wake,
				bath,
				bed,
				sleep_in,
				sleep,
				deep_sleep,
				description,
				created_at,
				updated_at
			) VALUES `
		var placeHolders []string
		var vals []interface{}
		for insertDayNum := dayNum; insertDayNum <= endDayNum; insertDayNum++ {
			placeHolders = append(placeHolders, "(?,?,?,?,?,?,?,?,?,?)")
			vals = append(
				vals,
				makeDayForInsert(year, monthNum, insertDayNum),
				0,
				0,
				0,
				"",
				"",
				"",
				"",
				now,
				now)
			dayNum++
		}
		insertQuery += strings.Join(placeHolders, ", ")
		result, err := db.Exec(insertQuery, vals...)
		if err != nil {
			log.Println("Insert month sleeps Error: ", err)
			return err
		}
		rows, _ := result.RowsAffected()
		log.Println("insert sleep suceed: ", rows)
	}
	return err
}
func makeDayForInsert(year int, month int, day int) time.Time {
	now := time.Now()
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, now.Location())
}
func updateSleep(sleepsFromFront []SleepFromFront) (sleeps []Sleep, err error) {

	// Date        string `json:"date"`
	// Wake        string `json:"wake"`
	// Bath        string `json:"bath"`
	// Bed         string `json:"bed"`
	// Sleep_in    string `json:"sleep_in"`
	// Sleep       string `json:"sleep"`
	// Deep_sleep  string `json:"deep_sleep"`
	// Description string `json:"description"`

	//	ID          int    `json:"id"`
	// Date        string `json:"date"`
	// DateStr     string `json:"date_str"`
	// Wake        int    `json:"wake"`
	// Bath        int    `json:"bath"`
	// Bed         int    `json:"bed"`
	// Sleep_in    string `json:"sleep_in"`
	// Sleep       string `json:"sleep"`
	// Deep_sleep  string `json:"deep_sleep"`
	// Description string `json:"description"`

	var updateSleeps []Sleep
	for _, sleepFromFront := range sleepsFromFront {
		var updateSleep Sleep
		updateSleep.ID = 0
		updateSleep.Date = sleepFromFront.Date
		updateSleep.Wake, err = strconv.Atoi(sleepFromFront.Wake)
		if err != nil {
			log.Println("Wake Conv Error: ", err)
			log.Println("Error Wake Str: ", sleepFromFront.Wake)
			return nil, err
		}
		updateSleep.Bath, err = strconv.Atoi(sleepFromFront.Bath)
		if err != nil {
			log.Println("Bath Conv Error: ", err)
			log.Println("Error Bath Str: ", sleepFromFront.Bath)
			return nil, err
		}
		updateSleep.Bed, err = strconv.Atoi(sleepFromFront.Bed)
		if err != nil {
			log.Println("Bed Conv Error: ", err)
			log.Println("Error Bed Str: ", sleepFromFront.Bed)
			return nil, err
		}
		updateSleep.Sleep_in = sleepFromFront.Sleep_in
		updateSleep.Sleep = sleepFromFront.Sleep
		updateSleep.Deep_sleep = sleepFromFront.Deep_sleep
		updateSleep.Description = sleepFromFront.Description
		updateSleeps = append(updateSleeps, updateSleep)
	}

	db, err := dbConnect()
	if err != nil {
		log.Println("DB Connect Error: ", err)
		return nil, err
	}
	defer func() {
		dbCloseErr := db.Close()
		if dbCloseErr != nil {
			log.Println("DB Close Error:", err)
			return
		}
		err = dbCloseErr
	}()
	updateQuery := `
		UPDATE sleeps 
		SET 
		  wake=?,
		  bath=?,
		  bed=?,
		  sleep_in=?,
		  sleep=?,
		  deep_sleep=?,
		  description=?,
		  updated_at=? 
		WHERE 
		  date=?`
	stmt, err := db.Prepare(updateQuery)
	if err != nil {
		log.Println("sleeps update error: ", err)
		return nil, err
	}
	defer func() {
		statementCloseErr := stmt.Close()
		if statementCloseErr != nil {
			log.Println("sleeps update Statement Close Error:", err)
			return
		}
		err = statementCloseErr
	}()

	now := time.Now()
	for _, updateSleep := range updateSleeps {
		_, err := stmt.Exec(
			updateSleep.Wake,
			updateSleep.Bath,
			updateSleep.Bed,
			updateSleep.Sleep_in,
			updateSleep.Sleep,
			updateSleep.Deep_sleep,
			updateSleep.Description,
			now,
			updateSleep.Date,
		)
		if err != nil {
			log.Println("sleeps update error: ", err)
			return nil, err
		}
	}
	// 2026-09-01の形式決め打ちで月引数を作ります。
	// 0123456789
	var tmpYear = updateSleeps[0].Date[:4]
	var tmpMonth = updateSleeps[0].Date[5:7]
	if len(tmpYear) != 4 {
		log.Println("Invalid YearStr: ", tmpYear)
		return nil, fmt.Errorf("Invalid YearStr: %v", tmpYear)
	}
	if len(tmpMonth) != 2 {
		log.Println("Invalid MonthStr: ", tmpMonth)
		return nil, fmt.Errorf("Invalid MonthStr: %v", tmpMonth)
	}
	var resultMonth = tmpYear + tmpMonth
	sleeps, err = getSleep(resultMonth)
	return sleeps, err
}
func changeDateString(dateStringFromDB string) (dateStringToDisp string) {
	return dateStringFromDB[:10]
}
func shapeMonth(monthFromURLQuery string) (month string, err error) {
	now := time.Now()
	if monthFromURLQuery == "" {
		month = now.Format("2006-01-02")
	} else {
		// yyyymmの形決め打ちで作成する。
		tmpYearStr := monthFromURLQuery[:4]
		tmpMonthStr := monthFromURLQuery[4:]
		if len(tmpYearStr) != 4 {
			log.Println("Invalid Year: ", tmpYearStr)
			return "", fmt.Errorf("Invalid Year: %v", tmpYearStr)
		}
		if len(tmpMonthStr) != 2 {
			log.Println("Invalid Month: ", tmpMonthStr)
			return "", fmt.Errorf("Invalid Month: %v", tmpMonthStr)
		}
		tmpYearNum, err := strconv.Atoi(tmpYearStr)
		if err != nil {
			log.Println("Year Conv Error: ", err)
			log.Println("Error Year Str: ", tmpYearStr)
			return "", fmt.Errorf("Invalid Year: %v", err)
		}
		tmpMonthNum, err := strconv.Atoi(tmpMonthStr)
		if err != nil {
			log.Println("Month Conv Error: ", err)
			log.Println("Error Month Str: ", tmpMonthStr)
			return "", fmt.Errorf("Invalid Month: %v", err)
		}
		if tmpMonthNum <= 0 || 12 < tmpMonthNum {
			log.Println("Invalid Month Error: ", tmpMonthNum)
			return "", fmt.Errorf("Invalid Month: %d", tmpMonthNum)
		}
		month = time.Date(tmpYearNum, time.Month(tmpMonthNum), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	}
	return month, nil
}
func getStartDay(month string) (startDay time.Time, err error) {
	now := time.Now()
	tmpMonth := month + " 00:00:00"
	monthDay, err := time.Parse("2006-01-02 15:04:05", tmpMonth)
	if err != nil {
		log.Println("first Day Parse Error:", err)
		log.Println("Fatal Date:", tmpMonth)
		return time.Time{}, err
	}
	firstDay := time.Date(monthDay.Year(), monthDay.Month(), 1, 0, 0, 0, 0, now.Location())
	return firstDay, nil
}
func getEndDay(month string) (startDay time.Time, err error) {
	now := time.Now()
	tmpMonth := month + " 23:59:59"
	monthDay, err := time.Parse("2006-01-02 15:04:05", tmpMonth)
	if err != nil {
		log.Println("last Day Parse Error:", err)
		log.Println("Fatal Date:", tmpMonth)
		return time.Time{}, err
	}
	lastDay := time.Date(monthDay.Year(), monthDay.Month(), 1, 23, 59, 59, 0, now.Location()).AddDate(0, 1, -1)
	return lastDay, nil
}
