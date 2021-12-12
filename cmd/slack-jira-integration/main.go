package main

import (
	"fmt"
	"net/http"

	"github.com/andygrunwald/go-jira"

	"github.com/spf13/viper"

	"github.com/gorilla/mux"
)

func main() {
	viper.BindEnv("USER_NAME")
	viper.BindEnv("PASSWORD")

	username := viper.GetString("USER_NAME")
	password := viper.GetString("PASSWORD")

	fmt.Println(fmt.Sprintf("username %s password %s", username, password))

	tp := jira.BasicAuthTransport{
		Username: username,
		Password: password,
	}

	jiraClient, _ := jira.NewClient(tp.Client(), "https://jordanshaw.atlassian.net/")

	r := runtime{
		JiraClient: jiraClient,
	}

	router := mux.NewRouter()
	router.HandleFunc("/slack/events", r.SlackEventsHandler)
	http.Handle("/", router)

	http.ListenAndServe(":8000", router)

}

type runtime struct {
	JiraClient *jira.Client
}

func (r *runtime) SlackEventsHandler(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(http.StatusOK)

}

func GetIssue(jiraClient jira.Client, issueID string) (*jira.Issue, error) {
	issue, _, err := jiraClient.Issue.Get(issueID, nil)

	return issue, err
}
