package slack

import(
    "testing"
	"github.com/golang/mock/gomock"
    "github.com/stretchr/testify/assert"
	"github.com/slack-go/slack"
)
func newMockSlacker(t *testing.T) *MockSlacker {
    ctrl := gomock.NewController(t)

    // Assert that Bar() is invoked.
    defer ctrl.Finish()

    return NewMockSlacker(ctrl)
}

func createSlackChannel() slack.Channel {
    return slack.Channel{
        GroupConversation: slack.GroupConversation{
            Name: "some-channel-name",
            Conversation: slack.Conversation{ 
                ID: "SOMECHANNELID",
            },
        },
    }

}

func createSlackMessage() slack.Message {
    return slack.Message{
        Msg: slack.Msg{
            Timestamp: "some-timestamp",
            Channel: "SOMECHANNELID",
            Text: "some message body",
        },

    }


}

func TestNewEnv(t *testing.T) {
    mockClient := newMockSlacker(t)

    expectedParams := slack.GetConversationsParameters{
        ExcludeArchived: true,
    }

    expectedChannel := createSlackChannel() 

    expectedEnv := &SlackEnv{
		SlackClient: mockClient,
		SlackSigningSecret: "some-secret",
        SlackChannelNames: []string{"some-channel-name"},
        SlackEmojis: map[string]string{"SOMECHANNELID": "some-emoji"},
        
	}


    mockClient.EXPECT().getConversations(&expectedParams).Times(1).Return([]slack.Channel{expectedChannel}, "", nil)

    env, err := NewEnv(mockClient, "some-secret", map[string]string{"some-channel-name":"some-emoji"}, []string{"some-channel-name"})

    assert.True(t, err == nil)
    assert.EqualValues(t, expectedEnv, env)


}

func TestGetConversationMessages(t *testing.T) {
    mockClient := newMockSlacker(t)

    env := &SlackEnv{
		SlackClient: mockClient,
		SlackSigningSecret: "some-secret",
        SlackChannelNames: []string{"some-channel-name"},
        SlackEmojis: map[string]string{"SOMECHANNELID": "some-emoji"},
        
	}
    expectedChannel := "some-channel"
    expectedTimestamp := "some-timestamp"

    expectedParams := slack.GetConversationRepliesParameters {
        ChannelID: expectedChannel,
        Timestamp: expectedTimestamp,
    }

    expectedMessage := createSlackMessage()

    mockClient.EXPECT().getConversationReplies(&expectedParams).Times(1).Return([]slack.Message{expectedMessage}, false, "", nil)

    messages, err := env.GetConversationMessages("some-channel", "some-timestamp")

    assert.True(t, err == nil)
    assert.EqualValues(t, messages, []slack.Message{expectedMessage})

}

func TestPostMessage(t *testing.T) {
    mockClient := newMockSlacker(t)

    env := &SlackEnv{
		SlackClient: mockClient,
		SlackSigningSecret: "some-secret",
        SlackChannelNames: []string{"some-channel-name"},
        SlackEmojis: map[string]string{"SOMECHANNELID": "some-emoji"},
        
	}

    mockClient.EXPECT().postMessage("SOMECHANNELID", "some-timestamp", "some-message-body").Times(1).Return("ok", "", nil)

    env.PostMessageToThread("SOMECHANNELID", "some-timestamp", "some-message-body")
}


