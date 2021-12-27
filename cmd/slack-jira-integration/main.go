package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/viper"

	"github.com/gorilla/mux"

    "slack-jira-integration/slack"
    "slack-jira-integration/jira"

	"github.com/slack-go/slack/slackevents"

)

func getEmojisByChannel(prefix string) map[string]string {
    emojis := make(map[string]string)
    for _, emoji := range emojis {
        emojisEnvVar := fmt.Sprintf("%s_%s", prefix, emoji)
        viper.BindEnv(emojisEnvVar)
        emojiFromEnv := viper.GetString(emojisEnvVar)
        emojis[emoji] = emojiFromEnv 
    }

    return emojis
}

func main() {
	viper.BindEnv("USER_NAME")
	viper.BindEnv("PASSWORD")
	viper.BindEnv("SLACK_SIGNING_SECRET")
	viper.BindEnv("SLACK_BOT_TOKEN")
	viper.BindEnv("JIRA_URL")
    viper.BindEnv("JIRA_PROJECT")
    viper.BindEnv("JIRA_SUMMARY")
    viper.BindEnv("JIRA_ISSUE_TYPE")

	username := viper.GetString("USER_NAME")
	password := viper.GetString("PASSWORD")

	slackSigningSecret := viper.GetString("SLACK_SIGNING_SECRET")
	slackBotToken := viper.GetString("SLACK_BOT_TOKEN")
    slackChannels := viper.GetStringSlice("SLACK_CHANNELS")
    slackEmojis := getEmojisByChannel("SLACK_EMOJI")

	jiraUrl := viper.GetString("JIRA_URL")
	jiraProject := viper.GetString("JIRA_PROJECT")
	jiraSummary := viper.GetString("JIRA_SUMMARY")
	jiraIssueType := viper.GetString("JIRA_ISSUE_TYPE")

    slackEnv := slack.NewEnv(slackBotToken, slackSigningSecret, slackEmojis, slackChannels)
    jiraEnv, err := jira.NewEnv(jiraUrl, username, password, jiraProject, jiraSummary, jiraIssueType)

    if err != nil {
        fmt.Println(fmt.Sprintf("jiraEnv err: %+v", err)) 
        return

    }

	r := runtime{        
        SlackEnv: slackEnv,
        JiraEnv: jiraEnv,
	}

	router := mux.NewRouter()
	router.Use(slack.ValidateSlackRequest(slackSigningSecret))
	router.HandleFunc("/slack/events", r.SlackEventsHandler)
	http.Handle("/", router)

	http.ListenAndServe(":8000", router)

}

type runtime struct {
    JiraEnv *jira.JiraEnv 
    SlackEnv *slack.SlackEnv
}



func (r *runtime) SlackEventsHandler(resp http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return

	}

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.ReactionAddedEvent:
          fmt.Println(fmt.Sprintf("ev: %+v", ev))
          r.ReactionAddedEvent(ev)
		

		}
	}

	resp.Write(body)

}



func (r *runtime) ReactionAddedEvent(ev *slackevents.ReactionAddedEvent) error {
    messages, err := r.SlackEnv.GetConversationMessages(ev.Item.Channel, ev.Item.Timestamp)

    if err != nil {
        return err
    }

    createdIssue, err := r.JiraEnv.CreateJiraIssue(messages[0].Msg.Text)

    if err != nil {
        return err
    }

    jiraUrlToIssue := fmt.Sprintf("%sbrowse/%s", r.JiraEnv.JiraUrl, createdIssue.Key)

    err = r.SlackEnv.PostMessageToThread(
        ev.Item.Channel,
        ev.Item.Timestamp,
        jiraUrlToIssue)

    return err

}


