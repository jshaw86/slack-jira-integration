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
    viper.BindEnv("JIRA_PROJECT")
    viper.BindEnv("JIRA_SUMMARY")
    viper.BindEnv("JIRA_ISSUE_TYPE")

	username := viper.GetString("USER_NAME")
	password := viper.GetString("PASSWORD")
	signingSecret := viper.GetString("SIGNING_SECRET")
	botToken := viper.GetString("BOT_TOKEN")

	jiraUrl := viper.GetString("JIRA_URL")
	jiraProject := viper.GetString("JIRA_PROJECT")
	jiraSummary := viper.GetString("JIRA_SUMMARY")
	jiraIssueType := viper.GetString("JIRA_ISSUE_TYPE")

	tp := jira.BasicAuthTransport{
		Username: username,
		Password: password,
	}

	jiraClient, _ := jira.NewClient(tp.Client(), jiraUrl)

	slackClient := slack.New(botToken)

    jiraUser, resp, err := jiraClient.User.GetSelf()

    if err != nil {
      bodyBytes, _ := ioutil.ReadAll(resp.Body)
      fmt.Println(fmt.Sprintf("jira fetch user err, can't start: %+v %+v", string(bodyBytes), err))
      return

    }

	r := runtime{
		JiraClient:    jiraClient,
		SlackClient:   slackClient,
		SigningSecret: signingSecret,
        JiraProject: jiraProject,
        JiraSummary: jiraSummary,
        JiraIssueType: jiraIssueType,
        JiraUserAccountID: jiraUser.AccountID,
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
    JiraProject string
    JiraSummary string
    JiraIssueType string
    JiraUserAccountID string
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

func (r *runtime) ReactionAddedEvent(ev *slackevents.ReactionAddedEvent) error {
    fmt.Println(fmt.Sprintf("ev %+v", ev))

    messages, err := GetConversationMessages(r.SlackClient, ev.Item.Channel, ev.Item.Timestamp)

    if err != nil {
        return err
    }

    issue := createJiraIssue(r.JiraProject, r.JiraIssueType, r.JiraSummary, messages[0].Msg.Text, r.JiraUserAccountID)

    createdIssue, _, err := r.JiraClient.Issue.CreateWithContext(context.Background(), issue)

    if err != nil {
        return err
    }

    fmt.Println(fmt.Sprintf("issue: %+v", createdIssue))

    r.SlackClient.PostMessage(ev.Item.Channel, slack.MsgOptionTS(ev.Item.Timestamp), slack.MsgOptionText(createdIssue.Key, true))

    return nil

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
