package main

import (
	"database/sql"
	"errors"
	"log"
	"strings"
)

var ErrNoFields = errors.New("no fields to update")

// loadActivities 函数从数据库中加载活动数据，并返回一个包含所有活动的切片（和error）。
func loadActivities(db *sql.DB) ([]Activity, error) { //db *sql.DB：把 main() 里的数据库连接池对象传进函数。(依旧指针形式)
	rows, err := db.Query(
		"SELECT id, name, capacity, open FROM activities",
	)
	if err != nil {
		return nil, err //查询数据库失败，返回错误
	}
	defer rows.Close()

	var activities []Activity

	for rows.Next() {
		var activity Activity

		err := rows.Scan(
			&activity.ID,
			&activity.Name,
			&activity.Capacity,
			&activity.Open,
		) //Scan() 方法把查询结果的每一行数据映射到 Activity 结构体的字段中,需要注意顺序必须相同
		if err != nil {
			return nil, err
		}

		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return activities, nil
}

// loadActivity()用于数据库（单个Activity）查询
func loadActivity(db *sql.DB, id int) (Activity, error) {
	var activity Activity

	//
	err := db.QueryRow(`
	SELECT id, name, capacity, open
	FROM activities
	WHERE id = ?
	`, id).Scan(
		&activity.ID,
		&activity.Name,
		&activity.Capacity,
		&activity.Open,
	)

	return activity, err
}

// 更新函数，负责更新数据库中的数据。RowsAffect()只用于观察修改情况，而不是直接认为活动不存在。需要更明确的错误模型。
func updateActivity(db *sql.DB, id int, req UpdateActivityRequest) error {
	//检查活动是否存在
	_, err := loadActivity(db, id)
	if err != nil {
		return err
	}

	var setClauses []string
	var args []any

	if req.Name != nil {
		setClauses = append(setClauses, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Capacity != nil {
		setClauses = append(setClauses, "capacity = ?")
		args = append(args, *req.Capacity)
	}
	if req.Open != nil {
		setClauses = append(setClauses, "open = ?")
		args = append(args, *req.Open)
	}

	//检查更改内容是否为空
	if len(setClauses) == 0 {
		return ErrNoFields
	}

	//填入需要更新的活动id
	args = append(args, id)

	query := `
	UPDATE activities
	SET ` + strings.Join(setClauses, ", ") + `
	WHERE id = ?
	`

	result, err := db.Exec(query, args...)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	log.Println("updated rows:", rows)

	return nil
}
