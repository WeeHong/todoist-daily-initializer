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
	Name    string `json:"name"`
	DueTime string `json:"due_time"`
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

func main() {
	lambda.Start(handleRequest)
}

func handleRequest() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	url := "https://api.todoist.com/sync/v8/sync"
	k := os.Getenv("TODOIST_KEY")
	b := "Bearer " + k

	location, _ := time.LoadLocation("Asia/Singapore")
	t := time.Now().In(location)

	items, _ := ioutil.ReadFile("task.json")
	var input []TaskList
	err = json.Unmarshal(items, &input)
	if err != nil {
		log.Fatalf("Failed to unmarshal json: %v", err)
	}

	var uuid1 *uuid.UUID
	var uuid2 *uuid.UUID
	var itemId *uuid.UUID
	var dd1 DueDate
	var dd2 DueDate
	var ap1 ArgumentProperty
	var ap2 ArgumentProperty
	var c []Commands
	var args TodoistRequest

	for _, s := range input {
		uuid1, _ = uuid.NewV4()
		uuid2, _ = uuid.NewV4()
		itemId, _ = uuid.NewV4()

		dd1 = DueDate{
			Lang:        "en",
			IsRecurring: false,
			String:      "every day",
			Date:        t.Format("2006-01-02"),
			Timezone:    "Asia/Singapore",
		}

		dd2 = DueDate{
			Lang:        "en",
			IsRecurring: false,
			String:      fmt.Sprintf("%s %s", t.Format("02 Jan"), s.DueTime),
			Date:        fmt.Sprintf("%sT%s:00", t.Format("2006-01-02"), s.DueTime),
			Timezone:    "Asia/Singapore",
		}

		ap1 = ArgumentProperty{
			ID:        uuid1.String(),
			Content:   s.Name,
			Due:       dd1,
			DateAdded: t.Format((time.RFC3339)),
			Priority:  1,
		}

		ap2 = ArgumentProperty{
			ID:     itemId.String() + "-reminder",
			ItemID: itemId.String(),
			Type:   "absolute",
			Due:    dd2,
		}

		c = []Commands{
			{
				Type:   "item_add",
				UUID:   uuid1.String(),
				TempID: itemId.String(),
				Args:   ap1,
			},
			{
				Type:   "reminder_add",
				UUID:   uuid2.String(),
				TempID: uuid2.String() + "-reminder-temp-id",
				Args:   ap2,
			},
		}
		args = TodoistRequest{
			SyncToken: "*",
			Commands:  c,
		}

		reqBody, err := json.Marshal(args)
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
}
