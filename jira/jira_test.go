package jira

import (
	"testing"

	"github.com/andygrunwald/go-jira"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func newMockJiraer(t *testing.T) *MockJiraer {
    ctrl := gomock.NewController(t)

    // Assert that Bar() is invoked.
    defer ctrl.Finish()

    return NewMockJiraer(ctrl)
}

func TestNewEnv(t *testing.T) {
    mockClient := newMockJiraer(t)
    expectedUser := jira.User{AccountID: "some-account-id"}
    expectedEnv := JiraEnv{
        JiraClient: mockClient, 
        JiraProject: "Some Project", 
        JiraSummary: "some summary", 
        JiraIssueType: "someIssueType", 
        JiraUserAccountID: "some-account-id",
    }
    mockClient.EXPECT().getSelf().Times(1).Return(&expectedUser, nil, nil)

    newEnv, err := NewEnv(mockClient,  "Some Project", "some summary", "someIssueType")

    assert.True(t, err == nil)
    assert.EqualValues(t, newEnv,&expectedEnv)

}

func TestCreateJiraIssue(t *testing.T) {
    mockClient := newMockJiraer(t)
    env := JiraEnv{
        JiraClient: mockClient, 
        JiraProject: "Some Project", 
        JiraSummary: "some summary", 
        JiraIssueType: "someIssueType", 
        JiraUserAccountID: "some-account-id",
    }

    expectedFields := &jira.IssueFields{
        Reporter: &jira.User{
            AccountID: env.JiraUserAccountID,
        },
        Description: "some description",
        Type: jira.IssueType{
            Name: env.JiraIssueType,
        },
        Project: jira.Project{
            Key: env.JiraProject,
        },
        Summary:  env.JiraSummary, 
    }

    expectedIssue := &jira.Issue{
        Fields: expectedFields, 
    }

    mockClient.EXPECT().createIssue(expectedIssue).Times(1).Return(expectedIssue, nil, nil)

    issue, err := env.CreateJiraIssue("some description")

    assert.True(t, err == nil)
    assert.EqualValues(t, issue, expectedIssue)

}
