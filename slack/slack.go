package slack

import (
    "fmt"
    "context"
	"github.com/slack-go/slack"
)

type SlackEnv struct {
	SlackClient   *slack.Client
	SlackSigningSecret string
    SlackChannels []string
    SlackEmojis map[string]string
}

func NewEnv(slackBotToken string, slackSigningSecret string, slackEmojis map[string]string, slackChannels []string) *SlackEnv {

	slackClient := slack.New(slackBotToken)

	return &SlackEnv{
		SlackClient:   slackClient,
		SlackSigningSecret: slackSigningSecret,
        SlackChannels: slackChannels,
        SlackEmojis: slackEmojis,
        
	}
}

func (s *SlackEnv) GetConversationMessages(channel string, timestamp string) ([]slack.Message, error) {
    params := slack.GetConversationRepliesParameters {
        ChannelID: channel,
        Timestamp: timestamp,
    }


    messages, hasMore, nextCursor, err := s.SlackClient.GetConversationRepliesContext(context.Background(), &params)

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

        messages, hasMore, nextCursor, err = s.SlackClient.GetConversationRepliesContext(context.Background(), &params)

        if err != nil {
           return conversationMessages, err

        }

        conversationMessages = append(conversationMessages, messages...)


    }

    return conversationMessages, nil 


}

func (s *SlackEnv) PostMessageToThread(channel string, timestamp string, msgBody string ) error {

    _, resp, err := s.SlackClient.PostMessage(channel, slack.MsgOptionTS(timestamp), slack.MsgOptionText(msgBody, true))

    if err != nil {
        return fmt.Errorf("post message failed err: %s", resp)
    }

    return nil

}
