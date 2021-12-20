package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/andygrunwald/go-jira"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	"github.com/spf13/viper"

	"github.com/gorilla/mux"
)

func main() {
	viper.BindEnv("USER_NAME")
	viper.BindEnv("PASSWORD")
	viper.BindEnv("SIGNING_SECRET")
	viper.BindEnv("BOT_TOKEN")
	viper.BindEnv("JIRA_URL")

	username := viper.GetString("USER_NAME")
	password := viper.GetString("PASSWORD")
	signingSecret := viper.GetString("SIGNING_SECRET")
	botToken := viper.GetString("BOT_TOKEN")
	jiraUrl := viper.GetString("JIRA_URL")

	fmt.Println(fmt.Sprintf("username %s password %s secret %s", username, password, signingSecret))

	tp := jira.BasicAuthTransport{
		Username: username,
		Password: password,
	}

	jiraClient, _ := jira.NewClient(tp.Client(), jiraUrl)

	slackClient := slack.New(botToken)

	r := runtime{
		JiraClient:    jiraClient,
		SlackClient:   slackClient,
		SigningSecret: signingSecret,
	}

	router := mux.NewRouter()
	router.Use(validateSlackRequest(signingSecret))
	router.HandleFunc("/slack/events", r.SlackEventsHandler)
	http.Handle("/", router)

	http.ListenAndServe(":8000", router)

}

type runtime struct {
	JiraClient    *jira.Client
	SlackClient   *slack.Client
	SigningSecret string
}

func validateSlackRequest(signingSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			bodyBytes, _ := ioutil.ReadAll(req.Body)
			req.Body.Close() //  must close
			req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

			sv, err := slack.NewSecretsVerifier(req.Header, signingSecret)
			if err != nil {
				resp.WriteHeader(http.StatusBadRequest)
				return
			}
			if _, err := sv.Write(bodyBytes); err != nil {
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}
			if err := sv.Ensure(); err != nil {
				resp.WriteHeader(http.StatusUnauthorized)
				return
			}
			eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(bodyBytes), slackevents.OptionNoVerifyToken())
			if err != nil {
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}

			if eventsAPIEvent.Type == slackevents.URLVerification {
				var r *slackevents.ChallengeResponse
				err := json.Unmarshal([]byte(bodyBytes), &r)
				if err != nil {
					resp.WriteHeader(http.StatusInternalServerError)
					return
				}
				resp.Header().Set("Content-Type", "text")
				resp.Write([]byte(r.Challenge))
				return
			}

			next.ServeHTTP(resp, req)

		})
	}
}

func (r *runtime) postBackMessage() {
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

	fmt.Println(fmt.Sprintf("outter: %+v", eventsAPIEvent))
	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.ReactionAddedEvent:
          r.ReactionAddedEvent(ev)
		

		}
	}

	resp.Write(body)

}

func createJiraIssue(issueProject string, issueType string, issueSummary string, description string, reporterAccountID string) *jira.Issue {
    fields := &jira.IssueFields{
        Reporter: &jira.User{
            AccountID: reporterAccountID,
        },
        Description: description,
        Type: jira.IssueType{
            Name: issueType,
        },
        Project: jira.Project{
            Key: issueProject,
        },
        Summary:  issueSummary, 
    }

    return &jira.Issue{
        Fields: fields, 
    }

}

func (r *runtime) ReactionAddedEvent(ev *slackevents.ReactionAddedEvent) {
    fmt.Println(fmt.Sprintf("ev %+v", ev))

    messages, err := GetConversationMessages(r.SlackClient, ev.Item.Channel, ev.Item.Timestamp)

    if err != nil {
        fmt.Println(fmt.Sprintf("get convo err: %+v", err))
        return
    }

    issue := createJiraIssue("TEST", "Story", "Slack Escalation", messages[0].Msg.Text, "61b50f96744c4d0069ad9201")

    createdIssue, resp, err := r.JiraClient.Issue.CreateWithContext(context.Background(), issue)

    if err != nil {
        body, err := ioutil.ReadAll(resp.Body)
        fmt.Println(fmt.Sprintf("err: %+v %+v %+v", err, resp.Response, string(body)))
        return
    }

    fmt.Println(fmt.Sprintf("createdIssues: %+v", createdIssue))

    r.SlackClient.PostMessage(ev.Item.Channel, slack.MsgOptionTS(ev.Item.Timestamp), slack.MsgOptionText("message received", true))

}

func GetIssue(jiraClient jira.Client, issueID string) (*jira.Issue, error) {
	issue, _, err := jiraClient.Issue.Get(issueID, nil)

	return issue, err
}

func GetConversationMessages(slackClient *slack.Client, channel string, timestamp string) ([]slack.Message, error) {
    params := slack.GetConversationRepliesParameters {
        ChannelID: channel,
        Timestamp: timestamp,
    }


    messages, hasMore, nextCursor, err := slackClient.GetConversationRepliesContext(context.Background(), &params)

    if err != nil {
        return nil, err

    }

    var conversationMessages []slack.Message 
    conversationMessages = append(conversationMessages, messages...)

    for hasMore {
        params := slack.GetConversationRepliesParameters {
            ChannelID: channel,
            Timestamp: timestamp,
            Cursor: nextCursor,
        }

        messages, hasMore, nextCursor, err = slackClient.GetConversationRepliesContext(context.Background(), &params)

        if err != nil {
           return conversationMessages, err

        }

        conversationMessages = append(conversationMessages, messages...)


    }

    return conversationMessages, nil 


}
