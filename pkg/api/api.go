package api

import (
	"encoding/json"

	"final_project/pkg/db"
	"final_project/pkg/rules"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const (
	Format = "20060102"
)

func Init() {
	http.HandleFunc("/api/nextdate", nextDayHandler)
	http.HandleFunc("/api/task", Auth(taskHandler))
	http.HandleFunc("/api/tasks", Auth(getTaskHandle))
	http.HandleFunc("/api/task/done", Auth(doneHandle))
	http.HandleFunc("/api/signin", SigninHandler)

}

func deleteHandle(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		ErrorJson(w, http.StatusBadRequest, "No id")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrorJson(w, http.StatusBadRequest, "cant parse string to int")
		return
	}

	err = db.DeleteTask(id)
	if err != nil {
		ErrorJson(w, http.StatusBadRequest, "cant delete task")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{})
}

func doneHandle(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		ErrorJson(w, http.StatusBadRequest, "cant update id")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrorJson(w, http.StatusBadRequest, "cant parse string to int")
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		ErrorJson(w, http.StatusBadRequest, "problems with finding task")
		return
	}

	if task.Repeat == "" {
		err = db.DeleteTask(id)
		if err != nil {
			ErrorJson(w, http.StatusBadRequest, "problems with deleting task")
			return
		}
	} else {

		date, _ := time.Parse(Format, task.Date)
		next, err := rules.NextDate(date, task.Date, task.Repeat)
		if err != nil {
			ErrorJson(w, http.StatusBadRequest, "problems with nextdate func")
			return
		}
		err = db.UpdateDate(next, id)
		if err != nil {
			ErrorJson(w, http.StatusBadRequest, "cant update date")
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{})
}
func putHandle(w http.ResponseWriter, r *http.Request) {
	var raw map[string]any
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		ErrorJson(w, http.StatusBadRequest, "cant read post")
		return
	}
	idStr, ok := raw["id"].(string)
	if !ok || idStr == "" {
		ErrorJson(w, http.StatusBadRequest, "id must be specified")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrorJson(w, http.StatusBadRequest, "wrong id format")
		return
	}
	title, ok := raw["title"].(string)
	if !ok || title == "" {
		ErrorJson(w, http.StatusBadRequest, "title must be specified")
		return
	}

	date, _ := raw["date"].(string)
	comment, _ := raw["comment"].(string)
	repeat, _ := raw["repeat"].(string)
	if date != "" {
		if _, err := time.Parse(Format, date); err != nil {
			ErrorJson(w, http.StatusBadRequest, "wrong date format")
			return
		}
	}
	if repeat != "" {
		if _, err := rules.NextDate(time.Now(), date, repeat); err != nil {
			ErrorJson(w, http.StatusBadRequest, "wrong repeat rule")
			return
		}
	}
	task := db.Task{
		ID:      id,
		Title:   title,
		Date:    date,
		Comment: comment,
		Repeat:  repeat,
	}
	if err := db.UpdateTask(&task); err != nil {
		ErrorJson(w, http.StatusBadRequest, "task not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{})
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		ErrorJson(w, http.StatusBadRequest, "cant read post")
		return
	}
	if task.Title == "" {
		ErrorJson(w, http.StatusBadRequest, "Title must be")
		return
	}
	now := time.Now()
	if task.Date == "" {
		task.Date = now.Format(Format)
	}
	dateParsed, err := time.Parse(Format, task.Date)
	if err != nil {
		ErrorJson(w, http.StatusBadRequest, "wrong date format")
		return
	}
	if now.Format(Format) > task.Date {
		if task.Repeat == "" {
			task.Date = now.Format(Format)
		} else {
			nextDay, _ := rules.NextDate(now, now.Format(Format), task.Repeat)
			task.Date = nextDay
		}
	}

	if err != nil {
		ErrorJson(w, http.StatusBadRequest, "wrong date format")
		return
	}
	if task.Repeat != "" {
		_, err := rules.NextDate(dateParsed, task.Date, task.Repeat)
		if err != nil {
			ErrorJson(w, http.StatusBadRequest, "problems with repeat rule")
			return
		}

	}

	id, err := db.AddTask(&task)
	if err != nil {
		ErrorJson(w, http.StatusBadRequest, "cant add task in db")
		return
	}
	task.ID = id

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func getHandle(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		ErrorJson(w, http.StatusBadRequest, "id cant be empty")
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		ErrorJson(w, http.StatusBadRequest, "cant parse number")
		return
	}
	ans, err := db.GetTask(idInt)
	if err != nil {
		ErrorJson(w, http.StatusBadRequest, "cant make get request")
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"id":      fmt.Sprint(ans.ID),
		"date":    ans.Date,
		"title":   ans.Title,
		"comment": ans.Comment,
		"repeat":  ans.Repeat,
	})
}

func getTaskHandle(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	ans, err := db.GetTasks(100, string(search))
	if err != nil {
		ErrorJson(w, http.StatusBadRequest, "cant make get request")
		return
	}

	w.Header().Set("Content-type", "application/json")

	ansMap := make([]map[string]string, 0, 40)
	for _, v := range ans {
		ansMap = append(ansMap, map[string]string{
			"id":      fmt.Sprint(v.ID),
			"date":    v.Date,
			"title":   v.Title,
			"comment": v.Comment,
			"repeat":  v.Repeat,
		})
	}
	json.NewEncoder(w).Encode(map[string]any{
		"tasks": ansMap,
	})
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getHandle(w, r)
	case http.MethodPut:
		putHandle(w, r)
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodDelete:
		deleteHandle(w, r)
	default:
		ErrorJson(w, http.StatusMethodNotAllowed, "there is no such method")
	}
}
func ErrorJson(w http.ResponseWriter, status int, mistake string) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": mistake,
	})
}

func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	date := r.FormValue("date")
	if date == "" {
		ErrorJson(w, http.StatusBadRequest, "date must be not empty")
		return
	}
	repeat := r.FormValue("repeat")
	nowstr := r.FormValue("now")
	var now time.Time
	if nowstr != "" {
		now, _ = time.Parse(Format, nowstr)
	} else {
		now = time.Now()
	}
	answ, err := rules.NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = w.Write([]byte(answ))
	if err != nil {
		ErrorJson(w, http.StatusBadRequest, "problems with writing")
		return

	}
}
