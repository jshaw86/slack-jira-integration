package slack

import (
    "fmt"
    "context"
	"github.com/slack-go/slack"
)

type SlackEnv struct {
	SlackClient Slacker
	SlackSigningSecret string
    SlackChannelNames []string
    SlackChannelIds []string
    SlackEmojis map[string]string
}

type Slacker interface {
    getConversationReplies(*slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error)
    getConversations(*slack.GetConversationsParameters) ([]slack.Channel, string, error)
    postMessage(string, string, string) (string, string, error)

}

type slackClient struct {
    Client *slack.Client
    Context context.Context

}

func NewClient(slackBotToken string) Slacker {
    context := context.Background()
    return &slackClient{
        Context: context,
        Client: slack.New(slackBotToken),
    }

}

func (s *slackClient) getConversations(params *slack.GetConversationsParameters) ([]slack.Channel, string, error) {
    return s.Client.GetConversationsContext(s.Context, params)
}

func (s *slackClient) getConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error) {
    return s.Client.GetConversationRepliesContext(s.Context, params)

}

func (s *slackClient) postMessage(channel string, timestamp string, msgBody string) (string, string, error)  {
    return s.Client.PostMessage(channel, slack.MsgOptionTS(timestamp), slack.MsgOptionText(msgBody, true))

}

func NewEnv(client Slacker, slackSigningSecret string, slackEmojis map[string]string, slackChannelNames []string) (*SlackEnv, error) {
    env := &SlackEnv{
		SlackClient: client,
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

func (s *SlackEnv) GetConversationMessages(channel string, timestamp string) ([]slack.Message, error) {
    params := slack.GetConversationRepliesParameters {
        ChannelID: channel,
        Timestamp: timestamp,
    }


    messages, hasMore, nextCursor, err := s.SlackClient.getConversationReplies(&params) 

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

        messages, hasMore, nextCursor, err = s.SlackClient.getConversationReplies(&params)

        if err != nil {
           return conversationMessages, err

        }

        conversationMessages = append(conversationMessages, messages...)

    }

    return conversationMessages, nil 

}

func (s *SlackEnv) PostMessageToThread(channel string, timestamp string, msgBody string ) error {
    _, resp, err := s.SlackClient.postMessage(channel, timestamp, msgBody) 

    if err != nil {
        return fmt.Errorf("post message failed err: %s", resp)
    }

    return nil

}

func (s *SlackEnv) getChannelID(channelName string) (string, error) {
    params := slack.GetConversationsParameters{
        ExcludeArchived: true,
    }

    channels, nextCursor, err := s.SlackClient.getConversations(&params) 

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

        channels, nextCursor, err = s.SlackClient.getConversations(&params) 

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

