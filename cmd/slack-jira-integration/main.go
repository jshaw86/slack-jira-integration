package main

import (
	"fmt"
    "strings"
	"net/http"

	"github.com/spf13/viper"

	"github.com/gorilla/mux"

    runtime "slack-jira-integration"
    "slack-jira-integration/slack"
    "slack-jira-integration/jira"

)

func getEmojisByChannel(prefix string, channels []string) map[string]string {
    emojis := make(map[string]string)
    for _, channel := range channels {
        emojisEnvVar := fmt.Sprintf("%s_%s", prefix, channel)
        viper.BindEnv(emojisEnvVar)
        emojiFromEnv := viper.GetString(emojisEnvVar)
        emojis[channel] = emojiFromEnv 
    }

    return emojis
}

func main() {
	viper.BindEnv("USER_NAME")
	viper.BindEnv("PASSWORD")
	viper.BindEnv("SLACK_SIGNING_SECRET")
	viper.BindEnv("SLACK_BOT_TOKEN")
    viper.BindEnv("SLACK_CHANNELS")
	viper.BindEnv("JIRA_URL")
    viper.BindEnv("JIRA_PROJECT")
    viper.BindEnv("JIRA_SUMMARY")
    viper.BindEnv("JIRA_ISSUE_TYPE")

	username := viper.GetString("USER_NAME")
	password := viper.GetString("PASSWORD")

	slackSigningSecret := viper.GetString("SLACK_SIGNING_SECRET")
	slackBotToken := viper.GetString("SLACK_BOT_TOKEN")
    slackChannels := strings.Split(viper.GetString("SLACK_CHANNELS"),",")
    slackEmojis := getEmojisByChannel("SLACK_EMOJI", slackChannels)

	jiraUrl := viper.GetString("JIRA_URL")
	jiraProject := viper.GetString("JIRA_PROJECT")
	jiraSummary := viper.GetString("JIRA_SUMMARY")
	jiraIssueType := viper.GetString("JIRA_ISSUE_TYPE")

    slackEnv, err := slack.NewEnv(slackBotToken, slackSigningSecret, slackEmojis, slackChannels)
    if err != nil {
        fmt.Println(fmt.Sprintf("slackEnv err: %+v", err)) 
        return

    }

    jiraEnv, err := jira.NewEnv(jiraUrl, username, password, jiraProject, jiraSummary, jiraIssueType)
    if err != nil {
        fmt.Println(fmt.Sprintf("jiraEnv err: %+v", err)) 
        return

    }

    r := runtime.New(slackEnv, jiraEnv)
	

    fmt.Println(fmt.Sprintf("env: %+v %+v", slackEnv, jiraEnv))

	router := mux.NewRouter()
	router.Use(slack.ValidateSlackRequest(slackSigningSecret))
	router.HandleFunc("/slack/events", r.SlackEventsHandler)
	http.Handle("/", router)

	http.ListenAndServe(":8000", router)

}


