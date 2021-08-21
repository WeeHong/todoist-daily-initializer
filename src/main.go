package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
	uuid "github.com/nu7hatch/gouuid"
)

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
	FullSync      bool     `json:"full_sync"`
	SyncStatus    response `json:"sync_status"`
	SyncToken     string   `json:"sync_token"`
	TempIDMapping response `json:"temp_id_mapping"`
}

type response map[string]string

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

	t := time.Now()

	var uuid1 *uuid.UUID
	var uuid2 *uuid.UUID
	var dd1 DueDate
	var dd2 DueDate
	var ap1 ArgumentProperty
	var ap2 ArgumentProperty
	var c []Commands
	var args TodoistRequest

	uuid1, _ = uuid.NewV4()
	uuid2, _ = uuid.NewV4()

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
		String:      t.Format("02 Jan") + " 20:00",
		Date:        t.Format("2006-01-02") + "T20:00:00",
		Timezone:    "Asia/Singapore",
	}

	ap1 = ArgumentProperty{
		ID:        "study-C#-course",
		Content:   "Study C# Course",
		Due:       dd1,
		DateAdded: t.Format((time.RFC3339)),
		Priority:  1,
	}

	ap2 = ArgumentProperty{
		ID:     "study-C#-course-reminder",
		ItemID: "study-C#-course",
		Type:   "absolute",
		Due:    dd2,
	}

	c = []Commands{
		{
			Type:   "item_add",
			UUID:   uuid1.String(),
			TempID: "study-C#-course",
			Args:   ap1,
		},
		{
			Type:   "reminder_add",
			UUID:   uuid2.String(),
			TempID: "study-C#-course-reminder-temp-id",
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

	err = json.NewDecoder(resp.Body).Decode(&TodoistResponse{})
	if err != nil {
		log.Fatalf("Error on decoding the response body: %v", err)
	}

	// ===========================================================================
	uuid1, _ = uuid.NewV4()
	uuid2, _ = uuid.NewV4()

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
		String:      t.Format("02 Jan") + " 22:00",
		Date:        t.Format("2006-01-02") + "T22:00:00",
		Timezone:    "Asia/Singapore",
	}

	ap1 = ArgumentProperty{
		ID:        "write-journal",
		Content:   "Write journal",
		Due:       dd1,
		DateAdded: t.Format((time.RFC3339)),
		Priority:  1,
	}

	ap2 = ArgumentProperty{
		ID:     "write-journal-reminder",
		ItemID: "write-journal",
		Type:   "absolute",
		Due:    dd2,
	}

	c = []Commands{
		{
			Type:   "item_add",
			UUID:   uuid1.String(),
			TempID: "write-journal",
			Args:   ap1,
		},
		{
			Type:   "reminder_add",
			UUID:   uuid2.String(),
			TempID: "write-journal-reminder-temp-id",
			Args:   ap2,
		},
	}

	args = TodoistRequest{
		SyncToken: "*",
		Commands:  c,
	}

	reqBody, err = json.Marshal(args)
	if err != nil {
		log.Fatalf("Failed to parse struct into JSON: %v", err)
	}
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", b)

	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("Error on response: %v", err)
	}
	defer resp.Body.Close()
}
