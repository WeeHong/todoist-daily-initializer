package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
	uuid "github.com/nu7hatch/gouuid"
)

type TaskList struct {
	Name             string `json:"name"`
	DueTime          string `json:"due_time"`
	CurrentDate      time.Time
	TaskID           string
	TaskDueDate      DueDate
	ReminderDueDate  DueDate
	TaskArgument     ArgumentProperty
	ReminderArgument ArgumentProperty
}

type TodoistRequest struct {
	SyncToken string     `json:"sync_token"`
	Commands  []Commands `json:"commands"`
}

type Commands struct {
	Type   string           `json:"type"`
	UUID   string           `json:"uuid"`
	TempID string           `json:"temp_id"`
	Args   ArgumentProperty `json:"args"`
}

type ArgumentProperty struct {
	ID        string  `json:"id"`
	ItemID    string  `json:"item_id"`
	Type      string  `json:"type"`
	Content   string  `json:"content"`
	Due       DueDate `json:"due"`
	DateAdded string  `json:"date_added"`
	Priority  int     `json:"priority"`
}

type DueDate struct {
	Lang        string `json:"lang"`
	IsRecurring bool   `json:"is_recurring"`
	String      string `json:"string"`
	Date        string `json:"date"`
	Timezone    string `json:"timezone"`
}

type TodoistResponse struct {
	FullSync      bool              `json:"full_sync"`
	SyncStatus    map[string]string `json:"sync_status"`
	SyncToken     string            `json:"sync_token"`
	TempIDMapping map[string]uint64 `json:"temp_id_mapping"`
}

func (i *TaskList) generateDueDate(t string) DueDate {
	var dd DueDate
	dd.Lang = "en"
	dd.IsRecurring = false
	dd.Timezone = "Asia/Singapore"
	dd.String = "every day"
	dd.Date = i.CurrentDate.Format("2006-01-02")

	if t == "reminder" {
		dd.String = fmt.Sprintf("%s %s", i.CurrentDate.Format("02 Jan"), i.DueTime)
		dd.Date = fmt.Sprintf("%sT%s:00", i.CurrentDate.Format("2006-01-02"), i.DueTime)
	}
	return dd
}

func (i *TaskList) generateArgument(t string) ArgumentProperty {
	uuid, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("Failed to generate UUID: %v", err)
	}
	var ap ArgumentProperty
	ap.ID = uuid.String()

	if t == "reminder" {
		ap.ItemID = i.TaskID
		ap.Type = "absolute"
		ap.Due = i.ReminderDueDate
	} else {
		i.TaskID = uuid.String()
		ap.Content = i.Name
		ap.DateAdded = i.CurrentDate.Format(time.RFC3339)
		ap.Due = i.TaskDueDate
		ap.Priority = 1
	}

	return ap
}

func (i *TaskList) generateCommand(t string) Commands {
	uuid, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("Failed to generate UUID: %v", err)
	}
	var c Commands

	if t == "reminder" {
		c.Type = "reminder_add"
		c.UUID = uuid.String()
		c.TempID = uuid.String()
		c.Args = i.ReminderArgument
	} else {
		c.Type = "item_add"
		c.UUID = uuid.String()
		c.TempID = i.TaskID
		c.Args = i.TaskArgument
	}

	return c
}

func (r *TodoistRequest) createItem() {

	url := "https://api.todoist.com/sync/v8/sync"
	k := os.Getenv("TODOIST_KEY")
	b := "Bearer " + k

	reqBody, err := json.Marshal(r)
	if err != nil {
		log.Fatalf("Failed to parse struct into JSON: %v", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", b)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error on response: %v", err)
	}
	defer resp.Body.Close()
	respReader, _ := ioutil.ReadFile("task.json")
	var output []TodoistResponse
	err = json.Unmarshal(respReader, &output)
	if err != nil {
		log.Fatalf("Failed to unmarshal json: %v", err)
	}
	fmt.Println(output)
}

func handleRequest() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var items []TaskList
	input := readJson()
	err = json.Unmarshal(input, &items)
	if err != nil {
		log.Fatalf("Failed to unmarshal json: %v", err)
	}

	location, _ := time.LoadLocation("Asia/Singapore")
	time := time.Now().In(location)

	var c []Commands

	for _, i := range items {
		i.CurrentDate = time

		i.TaskDueDate = i.generateDueDate("task")
		i.TaskArgument = i.generateArgument("task")

		if i.DueTime != "" {
			i.ReminderDueDate = i.generateDueDate("reminder")
			i.ReminderArgument = i.generateArgument("reminder")
		}

		c = []Commands{
			i.generateCommand("task"),
			i.generateCommand("reminder"),
		}

		args := TodoistRequest{
			SyncToken: "*",
			Commands:  c,
		}

		args.createItem()
	}
}

func readJson() []byte {
	items, err := ioutil.ReadFile("task.json")
	if err != nil {
		log.Fatalf("Failed to read task.json file: %v", err)
	}
	return items
}

func main() {
	lambda.Start(handleRequest)
}
