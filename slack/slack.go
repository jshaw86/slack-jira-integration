package slack

import (
    "fmt"
    "context"
	"github.com/slack-go/slack"
)

type SlackEnv struct {
    SlackContext context.Context
	SlackClient   *slack.Client
	SlackSigningSecret string
    SlackChannelNames []string
    SlackChannelIds []string
    SlackEmojis map[string]string
}

type SlackClientWrapper interface {
    getConversationReplies(*slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error)
    getConversations(*slack.GetConversationsParameters) ([]slack.Channel, string, error)

}

func NewEnv(slackBotToken string, slackSigningSecret string, slackEmojis map[string]string, slackChannelNames []string) (*SlackEnv, error) {
	slackClient := slack.New(slackBotToken)
    context := context.Background()

    env := &SlackEnv{
        SlackContext: context,
		SlackClient:   slackClient,
		SlackSigningSecret: slackSigningSecret,
        SlackChannelNames: slackChannelNames,
        
	}

    return env.setSlackEmojis(slackEmojis)

}

func (s *SlackEnv) setSlackEmojis(slackEmojis map[string]string) (*SlackEnv, error) {
    var slackChannelNamesToIds map[string]string
    for _, channelName := range s.SlackChannelNames {
        channelID, err := s.getChannelID(channelName)

        if err != nil {
            return nil, err

        }

        if err == nil && channelID == "" {
            return nil, fmt.Errorf("could not find channel name '%s'", channelName)

        }

        slackChannelNamesToIds[channelName] = channelID

    }

    slackEmojisByChannelID := make(map[string]string)

    for channelName, emoji := range slackEmojis {
        channelID := slackChannelNamesToIds[channelName]

        slackEmojisByChannelID[channelID] = emoji

    }

    s.SlackEmojis = slackEmojisByChannelID

    return s, nil

}

func (s *SlackEnv) getConversations(params *slack.GetConversationsParameters) ([]slack.Channel, string, error) {
    return s.SlackClient.GetConversationsContext(s.SlackContext, params)
}

func (s *SlackEnv) getConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error) {
    return s.SlackClient.GetConversationRepliesContext(s.SlackContext, params)

}


func (s *SlackEnv) GetConversationMessages(channel string, timestamp string) ([]slack.Message, error) {
    params := slack.GetConversationRepliesParameters {
        ChannelID: channel,
        Timestamp: timestamp,
    }


    messages, hasMore, nextCursor, err := s.getConversationReplies(&params) 

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

        messages, hasMore, nextCursor, err = s.getConversationReplies(&params)

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

func (s *SlackEnv) getChannelID(channelName string) (string, error) {
    params := slack.GetConversationsParameters{
        ExcludeArchived: true,
    }

    channels, nextCursor, err := s.getConversations(&params) 

    if err != nil {
        return "", err

    }

    for _, channel := range channels {

        if channel.Name == channelName {
            return channel.ID, nil

        }

    }

    for nextCursor != "" {
        params = slack.GetConversationsParameters{
            Cursor: nextCursor, 
            ExcludeArchived: true,
        }

        channels, nextCursor, err = s.getConversations(&params) 

        if err != nil {
            return "", err

        }

        for _, channel := range channels {

            if channel.Name == channelName {
                return channel.ID, nil

            }
        }
    }

    return "", nil

}

