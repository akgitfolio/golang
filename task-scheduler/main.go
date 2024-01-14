package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/robfig/cron/v3"
)

type Task struct {
	ID         int
	Schedule   string
	Command    string
	Args       []string
	LastRun    time.Time
	LastStatus string
}

var tasks []Task
var taskID int
var logFile *os.File

func main() {
	logFileName := flag.String("log", "task-scheduler.log", "Log file path")
	flag.Parse()

	var err error
	logFile, err = os.OpenFile(*logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	c := cron.New()
	scheduleTasks(c)
	c.Start()

	select {}
}

func scheduleTasks(c *cron.Cron) {
	addTask(c, "@every 1m", "echo", []string{"Hello, World!"})
	addTask(c, "0 0 * * *", "backup.sh", []string{"/path/to/backup"})
}

func addTask(c *cron.Cron, schedule, command string, args []string) {
	task := Task{
		ID:       taskID,
		Schedule: schedule,
		Command:  command,
		Args:     args,
	}
	taskID++
	tasks = append(tasks, task)

	entryID, err := c.AddFunc(schedule, func() {
		runTask(&task)
	})
	if err != nil {
		log.Printf("Failed to schedule task %d: %v", task.ID, err)
		return
	}
	task.ID = int(entryID)
}

func runTask(task *Task) {
	log.Printf("Running task %d: %s %v", task.ID, task.Command, task.Args)
	cmd := exec.Command(task.Command, task.Args...)
	err := cmd.Run()
	task.LastRun = time.Now()
	if err != nil {
		task.LastStatus = fmt.Sprintf("Failed: %v", err)
		log.Printf("Task %d failed: %v", task.ID, err)
	} else {
		task.LastStatus = "Success"
		log.Printf("Task %d completed successfully", task.ID)
	}
}
