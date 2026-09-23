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

type WakeTime struct {
	Wake int `json:"wake"`
}

type WakeTimeFromFront struct {
	WakeTime string `json:"wakeTime"`
}

type BathTime struct {
	Bath int `json:"bath"`
}

type BathTimeFromFront struct {
	BathTime string `json:"bathTime"`
}

type BedTime struct {
	Bed int `json:"bed"`
}

type BedTimeFromFront struct {
	BedTime string `json:"bedTime"`
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
			if err == nil {
				err = dbCloseErr
			}
		}
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
	endDay, err := getEndDay(month)
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
		sleep.DateStr, err = changeDateString(sleep.Date)
		if err != nil {
			log.Println("changeDate error: ", err)
			return nil, err
		}
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
		// wake_timeなどはデフォルト値に任せる。
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
			makedDay, err := makeDayForInsert(year, monthNum, insertDayNum)
			if err != nil {
				log.Print("make Day For Insert fail: ", err)
				return err
			}
			vals = append(
				vals,
				makedDay,
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
func makeDayForInsert(year int, month int, day int) (makedDayTime time.Time, err error) {
	if year <= 2000 {
		log.Print("Invalid year: ", year)
		return time.Time{}, fmt.Errorf("Invalid year num: %v", year)
	}
	if month <= 0 && 13 <= month {
		log.Print("Invalid month: ", month)
		return time.Time{}, fmt.Errorf("Invalid month num: %d", month)
	}
	if day <= 0 && 32 <= day {
		log.Print("Invalid day: ", day)
		return time.Time{}, fmt.Errorf("Invalid day num: %d", day)
	}
	now := time.Now()
	makedDayTime = time.Date(year, time.Month(month), day, 0, 0, 0, 0, now.Location())
	makedYear := makedDayTime.Year()
	makedMonth := int(makedDayTime.Month())
	makedDay := makedDayTime.Day()
	if year != makedYear || month != makedMonth || day != makedDay {
		log.Print("Invalid Pair of Year, month, Day: ", year, month, day)
		return time.Time{}, fmt.Errorf("Invalid Pair year, month, day: %d, %d, %d", year, month, day)
	}
	return makedDayTime, nil
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
			if err == nil {
				err = dbCloseErr
			}
		}
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
			if err == nil {
				err = statementCloseErr
			}
		}
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
func changeDateString(dateStringFromDB string) (dateStringToDisp string, err error) {
	if len(dateStringFromDB) != 25 {
		log.Println("Invalid date style: ", dateStringFromDB)
		return "", fmt.Errorf("Invalid date style: %v", dateStringFromDB)
	}
	dateStringToDisp = dateStringFromDB[:10]
	dateSlice := []rune(dateStringToDisp)
	if dateSlice[4] != '-' || dateSlice[7] != '-' {
		log.Println("Invalid date style: ", dateStringFromDB)
		return "", fmt.Errorf("Invalid date style: %v", dateStringFromDB)
	}
	return dateStringToDisp, nil
}
func shapeMonth(monthFromURLQuery string) (month string, err error) {
	now := time.Now()
	if monthFromURLQuery == "" {
		month = now.Format("2006-01-02")
	} else {
		// yyyymmの形決め打ちで作成する。
		if len(monthFromURLQuery) != 6 {
			log.Println("Invalid month style: ", monthFromURLQuery)
			return "", fmt.Errorf("Invalid month style: %v", monthFromURLQuery)
		}
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

func updateWake(updateTime WakeTimeFromFront) (wakeTime WakeTime, err error) {
	db, err := dbConnect()
	if err != nil {
		log.Println("db Connect Error: ", err)
		return wakeTime, err
	}
	defer func() {
		dbCloseErr := db.Close()
		if dbCloseErr != nil {
			log.Println("DB Close Error:", dbCloseErr)
			if err == nil {
				err = dbCloseErr
			}
		}
	}()
	wakeFixedTimeQuery := `
		SELECT
			FIXED_HOUR_TIME
		FROM
			FIXED_HOUR
	    WHERE
	    	FIXED_HOUR_NAME=?`
	wakeFixedTimeRows, err := db.Query(wakeFixedTimeQuery, "起床")
	if err != nil {
		log.Println("wake Fixed Time query error: ", err)
		return wakeTime, fmt.Errorf("wake Fixed Time query error: %v", err)
	}
	var wakeFixedTimeStr string
	for wakeFixedTimeRows.Next() {
		if err := wakeFixedTimeRows.Scan(&wakeFixedTimeStr); err != nil {
			log.Println("wake Fixed Time scan error: ", err)
			f := fmt.Errorf("wake Fixed Time scan error: %v", err)
			log.Print(f)
			if f != nil {
				log.Print("not null.")
			}
			return wakeTime, fmt.Errorf("wake Fixed Time scan error: %v", err)
		}
	}
	if err := wakeFixedTimeRows.Err(); err != nil {
		log.Println("wake Fixed Time rows error: ", err)
		return wakeTime, fmt.Errorf("wake Fixed Time rows error: %v", err)
	}

	wakeTimeStr := updateTime.WakeTime
	now := time.Now()

	// 起床時間なので0時またぎは対応しない。
	// 2026-09-22T13:32:00:00 決め打ち。
	// 01234567890123456
	if len(wakeTimeStr) != 22 {
		log.Print("Invalid Wake Fixed Time:", wakeTimeStr)
		return wakeTime, fmt.Errorf("Invalid Wake Fixed Time: %s", wakeTimeStr)
	}

	dayStr := wakeTimeStr[:10]
	log.Print(dayStr)
	wakeYearNum, err := strconv.Atoi(dayStr[:4])
	if err != nil {
		log.Print("Invalid year: ", err)
		return wakeTime, err
	}
	wakeMonthNum, err := strconv.Atoi(dayStr[5:7])
	if err != nil {
		log.Print("Invalid month: ", err)
		return wakeTime, err
	}
	wakeDayNum, err := strconv.Atoi(dayStr[8:10])
	if err != nil {
		log.Print("Invalid day: ", err)
		return wakeTime, err
	}
	wakeHourNum, err := strconv.Atoi(wakeTimeStr[11:13])
	if err != nil {
		log.Print("Invalid Hour: ", err)
		return wakeTime, err
	}
	wakeMinNum, err := strconv.Atoi(wakeTimeStr[14:16])
	if err != nil {
		log.Print("Invalid min: ", err)
		return wakeTime, err
	}

	wake := time.Date(wakeYearNum, time.Month(wakeMonthNum), wakeDayNum, wakeHourNum, wakeMinNum, 0, 0, now.Location())

	// wakeFixedTimeStr
	// 10:00:00
	// 01234567
	wakeFixedHourNum, err := strconv.Atoi(wakeFixedTimeStr[0:2])
	if err != nil {
		log.Print("Invalid fixed hour: ", err)
		return wakeTime, err
	}
	wakeFixedMinNum, err := strconv.Atoi(wakeFixedTimeStr[3:5])
	if err != nil {
		log.Print("Invalid fixed min: ", err)
		return wakeTime, err
	}

	wakeFixed := time.Date(wakeYearNum, time.Month(wakeMonthNum), wakeDayNum, wakeFixedHourNum, wakeFixedMinNum, 0, 0, now.Location())
	wakeTimeNum := wakeFixed.Sub(wake).Minutes()
	log.Println(wakeTimeNum)

	updateQuery := `
		UPDATE sleeps 
		SET 
		  wake=?,
		  wake_time=?,
		  updated_at=?		  
		WHERE 
		  date=?`
	stmt, err := db.Prepare(updateQuery)
	if err != nil {
		log.Println("wake update error: ", err)
		return wakeTime, err
	}
	defer func() {
		statementCloseErr := stmt.Close()
		if statementCloseErr != nil {
			log.Println("wake update Statement Close Error:", err)
			if err == nil {
				err = statementCloseErr
			}
		}
	}()

	_, err = stmt.Exec(wakeTimeNum, wake, now, dayStr)
	if err != nil {
		log.Println("sleeps update error: ", err)
		return wakeTime, err
	}
	wakeTime.Wake = int(wakeTimeNum)
	return wakeTime, nil
}

func updateBath(updateTime BathTimeFromFront) (bathTime BathTime, err error) {
	db, err := dbConnect()
	if err != nil {
		log.Println("db Connect Error: ", err)
		return bathTime, err
	}
	defer func() {
		dbCloseErr := db.Close()
		if dbCloseErr != nil {
			log.Println("DB Close Error:", dbCloseErr)
			if err == nil {
				err = dbCloseErr
			}
		}
	}()
	bathFixedTimeQuery := `
		SELECT
			FIXED_HOUR_TIME
		FROM
			FIXED_HOUR
	    WHERE
	    	FIXED_HOUR_NAME=?`
	bathFixedTimeRows, err := db.Query(bathFixedTimeQuery, "入浴")
	if err != nil {
		log.Println("bath Fixed Time query error: ", err)
		return bathTime, fmt.Errorf("bath Fixed Time query error: %v", err)
	}
	var bathFixedTimeStr string
	for bathFixedTimeRows.Next() {
		if err := bathFixedTimeRows.Scan(&bathFixedTimeStr); err != nil {
			log.Println("bath Fixed Time scan error: ", err)
			return bathTime, fmt.Errorf("bath Fixed Time scan error: %v", err)
		}
	}
	if err := bathFixedTimeRows.Err(); err != nil {
		log.Println("bath Fixed Time rows error: ", err)
		return bathTime, fmt.Errorf("bath Fixed Time rows error: %v", err)
	}

	bathTimeStr := updateTime.BathTime
	now := time.Now()

	// 0時またぎどうするかです。TODO。
	if len(bathTimeStr) != 22 {
		log.Print("Invalid bath Fixed Time:", bathTimeStr)
		return bathTime, fmt.Errorf("Invalid bath Fixed Time: %s", bathTimeStr)
	}

	dayStr := bathTimeStr[:10]
	bathYearNum, err := strconv.Atoi(dayStr[:4])
	if err != nil {
		log.Print("Invalid year: ", err)
		return bathTime, err
	}
	bathMonthNum, err := strconv.Atoi(dayStr[5:7])
	if err != nil {
		log.Print("Invalid month: ", err)
		return bathTime, err
	}
	bathDayNum, err := strconv.Atoi(dayStr[8:10])
	if err != nil {
		log.Print("Invalid day: ", err)
		return bathTime, err
	}
	bathHourNum, err := strconv.Atoi(bathTimeStr[11:13])
	if err != nil {
		log.Print("Invalid Hour: ", err)
		return bathTime, err
	}
	bathMinNum, err := strconv.Atoi(bathTimeStr[14:16])
	if err != nil {
		log.Print("Invalid min: ", err)
		return bathTime, err
	}

	bath := time.Date(bathYearNum, time.Month(bathMonthNum), bathDayNum, bathHourNum, bathMinNum, 0, 0, now.Location())

	// bathFixedTimeStr
	// 10:00:00
	// 01234567
	bathFixedHourNum, err := strconv.Atoi(bathFixedTimeStr[0:2])
	if err != nil {
		log.Print("Invalid fixed hour: ", err)
		return bathTime, err
	}
	bathFixedMinNum, err := strconv.Atoi(bathFixedTimeStr[3:5])
	if err != nil {
		log.Print("Invalid fixed min: ", err)
		return bathTime, err
	}

	bathFixed := time.Date(bathYearNum, time.Month(bathMonthNum), bathDayNum, bathFixedHourNum, bathFixedMinNum, 0, 0, now.Location())
	// 0時挟みの場合前日が定時の判定をします。
	if 0 <= bathHourNum && bathHourNum < 12 {
		bathFixed = bathFixed.AddDate(0, 0, -1)
		// bathYearNum
		// bathMonthNum
		// bathDayNum
		preDayNum := bathDayNum - 1
		dayStr = fmt.Sprintf("%04d-%02d-%02d", bathYearNum, bathMonthNum, preDayNum)
		log.Print(dayStr)
	}
	bathTimeNum := bathFixed.Sub(bath).Minutes()
	log.Println(bathTimeNum)

	updateQuery := `
		UPDATE sleeps 
		SET 
		  bath=?,
		  bath_time=?,
		  updated_at=?		  
		WHERE 
		  date=?`
	stmt, err := db.Prepare(updateQuery)
	if err != nil {
		log.Println("bath update error: ", err)
		return bathTime, err
	}
	defer func() {
		statementCloseErr := stmt.Close()
		if statementCloseErr != nil {
			log.Println("bath update Statement Close Error:", err)
			if err == nil {
				err = statementCloseErr
			}
		}
	}()

	_, err = stmt.Exec(bathTimeNum, bath, now, dayStr)
	if err != nil {
		log.Println("bath update error: ", err)
		return bathTime, err
	}
	bathTime.Bath = int(bathTimeNum)
	return bathTime, nil
}

func updateBed(updateTime BedTimeFromFront) (bedTime BedTime, err error) {
	db, err := dbConnect()
	if err != nil {
		log.Println("db Connect Error: ", err)
		return bedTime, err
	}
	defer func() {
		dbCloseErr := db.Close()
		if dbCloseErr != nil {
			log.Println("DB Close Error:", dbCloseErr)
			if err == nil {
				err = dbCloseErr
			}
		}
	}()
	bedFixedTimeQuery := `
		SELECT
			FIXED_HOUR_TIME
		FROM
			FIXED_HOUR
	    WHERE
	    	FIXED_HOUR_NAME=?`
	bedFixedTimeRows, err := db.Query(bedFixedTimeQuery, "就寝")
	if err != nil {
		log.Println("bed Fixed Time query error: ", err)
		return bedTime, fmt.Errorf("bed Fixed Time query error: %v", err)
	}
	var bedFixedTimeStr string
	for bedFixedTimeRows.Next() {
		if err := bedFixedTimeRows.Scan(&bedFixedTimeStr); err != nil {
			log.Println("bed Fixed Time scan error: ", err)
			return bedTime, fmt.Errorf("bed Fixed Time scan error: %v", err)
		}
	}
	if err := bedFixedTimeRows.Err(); err != nil {
		log.Println("bed Fixed Time rows error: ", err)
		return bedTime, fmt.Errorf("bed Fixed Time rows error: %v", err)
	}

	bedTimeStr := updateTime.BedTime
	now := time.Now()

	if len(bedTimeStr) != 22 {
		log.Print("Invalid bed Fixed Time:", bedTimeStr)
		return bedTime, fmt.Errorf("Invalid bed Fixed Time: %s", bedTimeStr)
	}

	dayStr := bedTimeStr[:10]
	bedYearNum, err := strconv.Atoi(dayStr[:4])
	if err != nil {
		log.Print("Invalid year: ", err)
		return bedTime, err
	}
	bedMonthNum, err := strconv.Atoi(dayStr[5:7])
	if err != nil {
		log.Print("Invalid month: ", err)
		return bedTime, err
	}
	bedDayNum, err := strconv.Atoi(dayStr[8:10])
	if err != nil {
		log.Print("Invalid day: ", err)
		return bedTime, err
	}
	bedHourNum, err := strconv.Atoi(bedTimeStr[11:13])
	if err != nil {
		log.Print("Invalid Hour: ", err)
		return bedTime, err
	}
	bedMinNum, err := strconv.Atoi(bedTimeStr[14:16])
	if err != nil {
		log.Print("Invalid min: ", err)
		return bedTime, err
	}

	bed := time.Date(bedYearNum, time.Month(bedMonthNum), bedDayNum, bedHourNum, bedMinNum, 0, 0, now.Location())

	// bedFixedTimeStr
	// 10:00:00
	// 01234567
	bedFixedHourNum, err := strconv.Atoi(bedFixedTimeStr[0:2])
	if err != nil {
		log.Print("Invalid fixed hour: ", err)
		return bedTime, err
	}
	bedFixedMinNum, err := strconv.Atoi(bedFixedTimeStr[3:5])
	if err != nil {
		log.Print("Invalid fixed min: ", err)
		return bedTime, err
	}

	bedFixed := time.Date(bedYearNum, time.Month(bedMonthNum), bedDayNum, bedFixedHourNum, bedFixedMinNum, 0, 0, now.Location())
	// 0時挟みの場合前日が定時の判定をします。
	if 0 <= bedHourNum && bedHourNum < 12 {
		bedFixed = bedFixed.AddDate(0, 0, -1)
		// bathYearNum
		// bathMonthNum
		// bathDayNum
		preDayNum := bedDayNum - 1
		dayStr = fmt.Sprintf("%04d-%02d-%02d", bedYearNum, bedMonthNum, preDayNum)
		log.Print(dayStr)
	}
	bedTimeNum := bedFixed.Sub(bed).Minutes()
	log.Println(bedTimeNum)

	updateQuery := `
		UPDATE sleeps 
		SET 
		  bed=?,
		  sleep_time=?,
		  updated_at=?		  
		WHERE 
		  date=?`
	stmt, err := db.Prepare(updateQuery)
	if err != nil {
		log.Println("bed update error: ", err)
		return bedTime, err
	}
	defer func() {
		statementCloseErr := stmt.Close()
		if statementCloseErr != nil {
			log.Println("bed update Statement Close Error:", err)
			if err == nil {
				err = statementCloseErr
			}
		}
	}()

	_, err = stmt.Exec(bedTimeNum, bed, now, dayStr)
	if err != nil {
		log.Println("bed update error: ", err)
		return bedTime, err
	}
	bedTime.Bed = int(bedTimeNum)
	return bedTime, nil
}
