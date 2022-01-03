package runtime 

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "strings"

    "github.com/golang/mock/gomock"
    "github.com/stretchr/testify/assert"

    "slack-jira-integration/slack"
    "slack-jira-integration/jira"
)

func newRuntime(t *testing.T) *runtime {
    ctrl := gomock.NewController(t)

    // Assert that Bar() is invoked.
    defer ctrl.Finish()

    mockJiraClient := jira.NewMockJiraer(ctrl)
    jiraEnv := &jira.JiraEnv{
        JiraClient: mockJiraClient, 
        JiraProject: "Some Project", 
        JiraSummary: "some summary", 
        JiraIssueType: "someIssueType", 
        JiraUserAccountID: "some-account-id",
    }

    mockSlackClient := slack.NewMockSlacker(ctrl)
    slackEnv := &slack.SlackEnv{
		SlackClient: mockSlackClient,
		SlackSigningSecret: "some-secret",
        SlackChannelNames: []string{"some-channel-name"},
        SlackEmojis: map[string]string{"SOMECHANNELID": "some-emoji"},
        
	}

    return &runtime{
        SlackEnv: slackEnv,
        JiraEnv: jiraEnv, 

    }

}

const (
    reactionAddedEventPayload = `{
        "token":"il5E6WOsnDv5rEcLcp0ftogS",
        "team_id":"TCJPJ3FAP",
        "api_app_id":"A02NU1FNPGX",
        "event":{
            "type":"reaction_added",
            "user":"UCJLPB2AG",
            "item":{"type":"message","channel":"CCK2APLUV","ts":"1641160687.000200"},
            "reaction":"white_check_mark",
            "item_user":"UCJLPB2AG",
            "event_ts":"1641160720.000300"
        },
        "type":"event_callback",
        "event_id":"Ev02SHMF0LR2",
        "event_time":1641160720,
        "authorizations":[{
            "enterprise_id":null,
            "team_id":"TCJPJ3FAP",
            "user_id":"UCJLPB2AG",
            "is_bot":false,
            "is_enterprise_install":false
        }],
        "is_ext_shared_channel":false,
        "event_context":"4-eyJldCI6InJlYWN0aW9uX2FkZGVkIiwidGlkIjoiVENKUEozRkFQIiwiYWlkIjoiQTAyTlUxRk5QR1giLCJjaWQiOiJDQ0syQVBMVVYifQ"
    }`
)


func TestSlackEventsHandler(t *testing.T) {
    // Create a request to pass to our handler. We don't have any query parameters for now, so we'll
    // pass 'nil' as the third parameter.
    req, err := http.NewRequest("POST", "/slack/events",strings.NewReader(reactionAddedEventPayload))
    if err != nil {
        t.Fatal(err)
    }

    r := newRuntime(t)

    // We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(r.SlackEventsHandler)

    // Our handlers satisfy http.Handler, so we can call their ServeHTTP method 
    // directly and pass in our Request and ResponseRecorder.
    handler.ServeHTTP(rr, req)

    assert.Equal(t, rr.Code, http.StatusOK)

    // Check the response body is what we expect.
    assert.Equal(t, rr.Body.String(), reactionAddedEventPayload)

}
