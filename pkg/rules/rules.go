package rules

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	Format = "20060102"
)

func dateWeekFunc(date time.Time) int {
	dateWeek := date.Weekday()
	if dateWeek == 0 {
		dateWeek = 7
	}
	return int(dateWeek)
}

func LeapYearFeb(date time.Time) bool {
	year := date.Year()

	if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
		if date.Month() == 2 && date.Day() == 29 {
			return true
		} else {
			return false
		}
	}
	return false
}

func afterNow(date, now time.Time) bool {
	return date.After(now)
}

func NextWeekDate(repeat string, dstr, now time.Time) (string, error) {
	repeatSplitDays := strings.Split(repeat, ",")
	var flag = false
	for !afterNow(dstr, now) && flag != true {

		for _, v := range repeatSplitDays {
			weekDay, err := strconv.Atoi(strings.TrimSpace(v))
			if err != nil {
				return "", fmt.Errorf("problem with parsing string type: %v", err)
			}
			if dateWeekFunc(dstr) == weekDay {
				flag = true

			}
		}
		if flag == false {
			dstr = dstr.AddDate(0, 0, 1)
		}

	}
	return dstr.Format(Format), nil

}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if len(repeat) == 0 {

		return "", nil
	}

	repeatSplit := strings.Split(repeat, " ")

	date, err := time.Parse(Format, dstart)
	if err != nil {
		return "", errors.New("problem with parsing time")

	}
	switch repeatSplit[0] {
	case "d":
		if len(repeatSplit) < 2 {
			return "", errors.New("Wrong repeat rule format")
		}
		count, err := strconv.Atoi(repeatSplit[1])

		if err != nil {
			return "", errors.New("Wrong repeat rule")
		}
		if count > 400 {
			return "", errors.New("days more than expected")
		}

		if date.After(now) {
			date = date.AddDate(0, 0, count)
		} else {
			for date.Before(now) || date.Equal(now) {
				date = date.AddDate(0, 0, count)
			}
		}
		return date.Format(Format), nil

	case "y":
		var counterYears int
		if len(repeatSplit) == 1 {
			counterYears = 1
		} else if len(repeatSplit) == 2 {
			counterYears, err = strconv.Atoi(repeatSplit[1])
			if err != nil {
				return "", fmt.Errorf("problem with parsing %v", err)
			}
		} else {
			return "", errors.New("wrong format")
		}

		if LeapYearFeb(date) {
			date = date.AddDate(counterYears, 0, 0)
			//date = date.AddDate(0, 0, 1)
		} else {
			date = date.AddDate(counterYears, 0, 0)
		}
		if !date.After(now) {
			for !date.After(now) {
				if LeapYearFeb(date) {
					date = date.AddDate(counterYears, 0, 0)
					//date = date.AddDate(0, 0, 1)
				} else {
					date = date.AddDate(counterYears, 0, 0)
				}
			}
		}

		return date.Format(Format), nil

	case "w":
		return "", errors.New("Dont have w part")
	case "m":
		// Сделать позже если будет время
		return "", errors.New("Dont have m part")

	default:
		return "", errors.New("problems with indification repeat rule")
	}

}
