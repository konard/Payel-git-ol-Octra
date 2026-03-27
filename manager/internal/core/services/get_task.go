package services

import (
	"log"
	"manager/pkg/database"
)

func GetTask(taskId string) {
	task := database.Db.Find(&taskId, "task_id = ?", taskId)
	if task.Error != nil {
		log.Println("Task not found" + task.Error.Error())
	}

}
