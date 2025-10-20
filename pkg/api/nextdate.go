package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const DateFormat = "20060102"

// afterNow возвращает true, если d > now (только по дате, без учёта времени)
func afterNow(d, now time.Time) bool {
	y1, m1, d1 := d.Date()
	y2, m2, d2 := now.Date()
	if y1 > y2 {
		return true
	}
	if y1 == y2 && m1 > m2 {
		return true
	}
	if y1 == y2 && m1 == m2 && d1 > d2 {
		return true
	}
	return false
}

// parseIntList парсит строку с числами через запятую
func parseIntList(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		num, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		result = append(result, num)
	}
	return result, nil
}

// NextDate вычисляет следующую дату задачи по правилам повторения
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("repeat rule is empty")
	}
	date, err := time.Parse(DateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("invalid start date: %w", err)
	}

	parts := strings.Split(repeat, " ")
	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return "", errors.New("invalid format for d rule")
		}
		interval, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", errors.New("invalid interval in d rule")
		}
		if interval < 1 || interval > 400 {
			return "", errors.New("d interval must be between 1 and 400")
		}
		// Всегда добавляем интервал хотя бы раз, затем продолжаем пока date <= now
		for {
			date = date.AddDate(0, 0, interval)
			if afterNow(date, now) {
				break
			}
		}

	case "y":
		origMonth := date.Month()
		origDay := date.Day()
		// Всегда добавляем год хотя бы раз, затем продолжаем пока date <= now
		for {
			year := date.Year() + 1
			newDate := time.Date(year, origMonth, origDay, 0, 0, 0, 0, date.Location())
			if newDate.Month() != origMonth {
				// Если дата сдвинулась (29 февраля в невисокосном году), берем 1 марта
				newDate = time.Date(year, time.March, 1, 0, 0, 0, 0, date.Location())
			}
			date = newDate
			if afterNow(date, now) {
				break
			}
		}

	case "w":
		if len(parts) != 2 {
			return "", errors.New("invalid format for w rule")
		}
		weekdays, err := parseIntList(parts[1])
		if err != nil {
			return "", errors.New("invalid weekday list in w rule")
		}
		if len(weekdays) == 0 {
			return "", errors.New("weekday list is empty in w rule")
		}
		// Проверяем корректность дней недели (1-7)
		for _, wd := range weekdays {
			if wd < 1 || wd > 7 {
				return "", errors.New("weekday must be between 1 and 7")
			}
		}
		// Создаём карту допустимых дней недели
		validWeekdays := make(map[time.Weekday]bool)
		for _, wd := range weekdays {
			// Преобразуем 1-7 (понедельник-воскресенье) в time.Weekday (0-6, воскресенье-суббота)
			var weekday time.Weekday
			if wd == 7 {
				weekday = time.Sunday
			} else {
				weekday = time.Weekday(wd)
			}
			validWeekdays[weekday] = true
		}
		// Ищем следующую подходящую дату
		for {
			date = date.AddDate(0, 0, 1)
			if afterNow(date, now) && validWeekdays[date.Weekday()] {
				break
			}
		}

	case "m":
		if len(parts) < 2 {
			return "", errors.New("invalid format for m rule")
		}
		days, err := parseIntList(parts[1])
		if err != nil {
			return "", errors.New("invalid day list in m rule")
		}
		if len(days) == 0 {
			return "", errors.New("day list is empty in m rule")
		}
		// Проверяем корректность дней месяца
		for _, d := range days {
			if d == -3 || d < -2 || d == 0 || d > 31 {
				return "", errors.New("day must be between 1 and 31, or -1, -2")
			}
		}

		var months []int
		if len(parts) >= 3 {
			months, err = parseIntList(parts[2])
			if err != nil {
				return "", errors.New("invalid month list in m rule")
			}
			// Проверяем корректность месяцев
			for _, m := range months {
				if m < 1 || m > 12 {
					return "", errors.New("month must be between 1 and 12")
				}
			}
		}

		// Создаём карты допустимых дней и месяцев
		validDays := make(map[int]bool)
		for _, d := range days {
			validDays[d] = true
		}

		validMonths := make(map[int]bool)
		if len(months) > 0 {
			for _, m := range months {
				validMonths[m] = true
			}
		} else {
			// Если месяцы не указаны, все месяцы допустимы
			for m := 1; m <= 12; m++ {
				validMonths[m] = true
			}
		}

		// Ищем следующую подходящую дату
		for {
			date = date.AddDate(0, 0, 1)
			if !afterNow(date, now) {
				continue
			}

			month := int(date.Month())
			if !validMonths[month] {
				continue
			}

			day := date.Day()

			// Проверяем положительные дни
			if validDays[day] {
				break
			}

			// Проверяем отрицательные дни (-1, -2)
			if validDays[-1] || validDays[-2] {
				// Находим последний день месяца
				lastDay := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, date.Location()).Day()
				if validDays[-1] && day == lastDay {
					break
				}
				if validDays[-2] && day == lastDay-1 {
					break
				}
			}
		}

	default:
		return "", errors.New("unsupported repeat rule")
	}

	return date.Format(DateFormat), nil
}

func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var now time.Time
	var err error

	if nowStr == "" {
		now = time.Now()
	} else {
		now, err = time.Parse(DateFormat, nowStr)
		if err != nil {
			http.Error(w, "invalid now date format", http.StatusBadRequest)
			return
		}
	}

	next, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(next))
}
