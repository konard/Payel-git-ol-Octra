package services

import (
	"log"
	"manager/pkg/database"
)

func GetTask(taskId string) interface{} {
	var task interface{}
	result := database.Db.Where("task_id = ?", taskId).First(&task)
	if result.Error != nil {
		log.Println("Task not found: " + result.Error.Error())
		return nil
	}
	return task
}
