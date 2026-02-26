package db

import (
	"database/sql"

	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

const (
	schema = `create table scheduler(
				id integer primary key autoincrement, 
				date char(8) not null default "", 
				title VARCHAR, 
				comment text, 
				repeat VARCHAR
			)`
)
const (
	Format = "20060102"
)

type Task struct {
	ID      int    `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

var (
	install bool
	Db      *sql.DB
)

func DeleteTask(id int) error {
	_, err := Db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	return err
}

func UpdateDate(next string, id int) error {
	res, err := Db.Exec("UPDATE scheduler SET date = ? WHERE id = ?", next, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("incorrect id for updating date")
	}
	return nil
}

func UpdateTask(t *Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := Db.Exec(query, t.Date, t.Title, t.Comment, t.Repeat, t.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("incorrect id for updating task")
	}
	return nil
}

func GetTask(id int) (Task, error) {

	task := Task{}
	err := Db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

func GetTasks(limit int, search string) ([]*Task, error) {
	ans := []*Task{}
	query := "SELECT id, date, title, comment, repeat FROM scheduler "
	args := []any{}
	if search != "" {
		if t, err := time.Parse("02.01.2006", search); err == nil {
			query += "WHERE date = ? "
			args = append(args, t.Format(Format))
		} else {
			query += "where title like ? or comment like ? "
			like := "%" + search + "%"
			args = append(args, like, like)
		}

	}
	query += "ORDER BY date ASC LIMIT ?"
	args = append(args, int64(limit))
	rows, err := Db.Query(query, args...)
	if err != nil {
		return []*Task{}, err
	}
	defer rows.Close()

	for rows.Next() {
		task := Task{}
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

		if err != nil {
			return nil, err
		}

		ans = append(ans, &task)

	}
	if err = rows.Err(); err != nil {
		return []*Task{}, err
	}
	if ans == nil {
		return []*Task{}, nil
	}

	return ans, err
}

func AddTask(t *Task) (int, error) {
	line := "insert into scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"
	res, err := Db.Exec(line, t.Date, t.Title, t.Comment, t.Repeat)
	if err != nil {
		return 0, fmt.Errorf("problems with sql command %v", err)

	}
	resId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(resId), nil
}

func Init(dbFile string) error {
	var err error
	_, err = os.Stat(dbFile)
	if err != nil {
		install = true
	}

	Db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	if install {
		_, err := Db.Exec(schema)
		if err != nil {
			return err
		}
		_, err = Db.Exec(`CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date)`)
		if err != nil {
			return err
		}
	}
	return nil

}
