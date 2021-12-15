package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/andygrunwald/go-jira"

	"github.com/spf13/viper"

	"github.com/gorilla/mux"

	"github.com/coro/verifyslack"
)

func main() {
	viper.BindEnv("USER_NAME")
	viper.BindEnv("PASSWORD")
	viper.BindEnv("SIGNING_SECRET")

	username := viper.GetString("USER_NAME")
	password := viper.GetString("PASSWORD")
	signingSecret := viper.GetString("SIGNING_SECRET")

	fmt.Println(fmt.Sprintf("username %s password %s secret %s", username, password, signingSecret))

	tp := jira.BasicAuthTransport{
		Username: username,
		Password: password,
	}

	jiraClient, _ := jira.NewClient(tp.Client(), "https://jordanshaw.atlassian.net/")

	r := runtime{
		JiraClient:    jiraClient,
		SigningSecret: signingSecret,
	}

	var validationTime time.Time
	validationTimeGetter := validationTimeGetter{validationTime: validationTime}

	router := mux.NewRouter()
	router.HandleFunc("/slack/events", verifyslack.RequestHandler(r.SlackEventsHandler, validationTimeGetter, signingSecret))
	http.Handle("/", router)

	http.ListenAndServe(":8000", router)

}

type runtime struct {
	JiraClient    *jira.Client
	SigningSecret string
}

type validationTimeGetter struct {
	validationTime time.Time
}

func (v validationTimeGetter) Now() time.Time {
	return v.validationTime
}

func slackSignature(timestamp string, requestBody []byte, signingSecret string) string {
	baseSignature := append([]byte(fmt.Sprintf("v0:%s:", timestamp)), requestBody...)
	mac := hmac.New(sha256.New, []byte(signingSecret))
	mac.Write(baseSignature)

	expectedSignature := fmt.Sprintf("v0=%s", hex.EncodeToString(mac.Sum(nil)))
	return expectedSignature
}

func (r *runtime) SlackEventsHandler(resp http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return

	}
	/*
		timestamp := req.Header.Get("X-Slack-Request-Timestamp")
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte(err.Error()))

		}
		givenSlackSignature := req.Header.Get("X-Slack-Signature")

		calcSlackSignature := slackSignature("v0", timestamp, string(body), r.SigningSecret)

		if givenSlackSignature != calcSlackSignature {
			fmt.Println(fmt.Errorf("calc %s given %s", calcSlackSignature, givenSlackSignature))
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte("signatures don't match"))

		}
	*/

	resp.Write(body)

}

func GetIssue(jiraClient jira.Client, issueID string) (*jira.Issue, error) {
	issue, _, err := jiraClient.Issue.Get(issueID, nil)

	return issue, err
}
