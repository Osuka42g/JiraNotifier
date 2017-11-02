package main

import (
	"fmt"
	"log"
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

type jiraIssue struct {
	*jira.Issue
}

type notification struct {
	body     string
	title    string
	subtitle string
	link     string
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

				n := newNotificationFromIssue(issue)
				fmt.Println(issue.Key, "has some changes ->", jiraIssue{&issue}.getLink(), "at", timeNow())

				go func(n notification) {
					sendNotification(n)
					time.Sleep(5 * time.Second)
				}(n)
			}
		}
		time.Sleep((1 * time.Minute) - (1 * time.Second))
	}
}

func sendNotification(n notification) {
	note := gosxnotifier.NewNotification(n.body)
	note.Title = n.title
	note.Subtitle = n.subtitle
	note.Sound = gosxnotifier.Blow
	note.Group = "com.osuka42g.JiraNotifier"
	note.Link = n.link
	note.AppIcon = "assets/jiraicon.ico"
	err := note.Push()

	if err != nil {
		log.Println(err)
	}
}

func newNotificationFromIssue(i jira.Issue) notification {
	return notification{
		body:     "You have new activity",
		title:    i.Key,
		subtitle: i.Fields.Summary,
		link:     jiraIssue{&i}.getLink(),
	}
}

func (ji jiraIssue) getLink() string {
	return os.Getenv("JIRA_HOST") + "browse/" + ji.Key
}

func timeNow() string {
	t := time.Now()
	return t.Format("15:04:05")
}
