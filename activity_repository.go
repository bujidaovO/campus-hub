package main

import "database/sql"

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

// loadActivity()用于数据库（单个）查询
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
