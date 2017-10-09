package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/deckarep/gosx-notifier"
	"github.com/joho/godotenv"
)

type jiraConfig struct {
	host     string
	username string
	password string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	jc := jiraConfig{host: os.Getenv("JIRA_HOST"), username: os.Getenv("JIRA_USERNAME"), password: os.Getenv("JIRA_PASSWORD")}

	jiraClient, err := jira.NewClient(nil, jc.host)
	if err != nil {
		panic(err)
	}
	res, err := jiraClient.Authentication.AcquireSessionCookie(jc.username, jc.password)
	if err != nil || res == false {
		fmt.Printf("Result: %v\n", res)
		panic(err)
	}

	jql := "PROJECT = WIZE and UPDATED > '-1m' and WATCHER = '" + jc.username + "'"

	for {
		issues, _, _ := jiraClient.Issue.Search(jql, nil)
		if len(issues) > 0 {
			for _, issue := range issues {
				sendNotification("You have new activity", issue.Key, issue.Fields.Summary, jc.host+"browse/"+issue.Key)
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

func checkLink(link string, c chan string) {
	_, err := http.Get(link)
	if err != nil {
		fmt.Println(link, "might be down!")
		c <- link
		return
	}
	fmt.Println(link, "is up!")
	c <- link
}

func sendNotification(action string, ticket string, title string, link string) {
	note := gosxnotifier.NewNotification(action)
	note.Title = ticket
	note.Subtitle = title
	note.Sound = gosxnotifier.Blow
	note.Group = "com.osuka42g.JiraNotifier"
	note.Link = link
	note.AppIcon = "assets/jiraicon.ico"
	err := note.Push()

	if err != nil {
		log.Println(err)
	}
}
