package runtime

import (
    "fmt"
	"io/ioutil"
	"encoding/json"
	"net/http"

    "slack-jira-integration/slack"
    "slack-jira-integration/jira"

	"github.com/slack-go/slack/slackevents"
)

type runtime struct {
    JiraEnv *jira.JiraEnv 
    SlackEnv *slack.SlackEnv
}

// New, create a new runtime, given a SlackEnv and JiraEnv
func New(slackEnv *slack.SlackEnv, jiraEnv *jira.JiraEnv) *runtime {
    return &runtime{        
        SlackEnv: slackEnv,
        JiraEnv: jiraEnv,
    }

}

// SlackEventsHandler, main server handler accepts requests from Slack client
// and routes Slack event type to right function
func (r *runtime) SlackEventsHandler(resp http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
    fmt.Println(string(body))
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
            r.reactionAddedEvent(ev)
        default:
            fmt.Println(fmt.Sprintf("ev: %+v", ev))
        }
	}

	resp.Write(body)

}

// channelEmojiCombinationMatches, logic to determine if given channelID and reaction(emoji)
// should be processed
func (r *runtime) channelEmojiCombinationMatches(channelID string, reaction string) bool {
    emoji, exists := r.SlackEnv.SlackEmojis[channelID]
    return exists && reaction == emoji

}

// reactionAddedEvent, handle a ReactionAddedEvent(emoji added) to top level thread
func (r *runtime) reactionAddedEvent(ev *slackevents.ReactionAddedEvent) error {
    // noop if channel and reaction do not exist or match desired channel/emoji combination
    if !r.channelEmojiCombinationMatches(ev.Item.Channel, ev.Reaction) {
        return nil
    }

    // get all messages in the current conversation
    messages, err := r.SlackEnv.GetConversationMessages(ev.Item.Channel, ev.Item.Timestamp)

    if err != nil {
        return err
    }

    // create a Jira issue with the text of the first message in the thread
    createdIssue, err := r.JiraEnv.CreateJiraIssue(messages[0].Msg.Text)

    if err != nil {
        return err
    }

    jiraUrlToIssue := fmt.Sprintf("%sbrowse/%s", r.JiraEnv.JiraUrl, createdIssue.Key)

    // post back to slack thread with link to jira issue created
    err = r.SlackEnv.PostMessageToThread(
        ev.Item.Channel,
        ev.Item.Timestamp,
        jiraUrlToIssue)

    return err

}

