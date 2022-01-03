package jira

import (
    "context"
    "io/ioutil"
	"github.com/andygrunwald/go-jira"
    "fmt"
)

// JiraEnv is the Dependency Injection(DI) for jira object allowing environment
// variable access and better testability
type JiraEnv struct {
	JiraClient Jiraer 
    JiraUrl string
    JiraProject string
    JiraSummary string
    JiraIssueType string
    JiraUserAccountID string
}

// Jiraer is an interface for testing purposes wrapping the concrete jira.client
// the JiraEnv takes a Jiraer which is either the jiraClient struct below or a 
// MockJiraClient from jira_test
type Jiraer interface {
    getSelf() (*jira.User, *jira.Response, error)
    createIssue(*jira.Issue) (*jira.Issue, *jira.Response, error)
}

type jiraClient struct{
    Context context.Context
    Client *jira.Client
}

// NewClient, construct a newClient which implements Jiraer 
func NewClient(jiraUrl string, username string, password string) Jiraer{
    tp := jira.BasicAuthTransport{
        Username: username,
        Password: password,
    }

    client, _ := jira.NewClient(tp.Client(), jiraUrl)
    return &jiraClient{Context: context.Background(), Client: client}

}

func (j *jiraClient) createIssue(issue *jira.Issue) (*jira.Issue, *jira.Response, error) {
    return j.Client.Issue.CreateWithContext(j.Context, issue)

}

func (j *jiraClient) getSelf() (*jira.User, *jira.Response, error) {
    return j.Client.User.GetSelf()

}

func (j *JiraEnv) setAccountID(accountID string) *JiraEnv {
    j.JiraUserAccountID = accountID
    return j
}

func NewEnv(client Jiraer, jiraProject string, jiraSummary string, jiraIssueType string) (*JiraEnv, error) {
    env := &JiraEnv{
        JiraClient: client,
        JiraProject: jiraProject,
        JiraSummary: jiraSummary,
        JiraIssueType: jiraIssueType,
    }

    jiraUser, resp, err := env.JiraClient.getSelf()

    if err != nil {
        bodyBytes, _ := ioutil.ReadAll(resp.Body)
        return nil, fmt.Errorf("get self err: %s", bodyBytes)

    }

    return env.setAccountID(jiraUser.AccountID), nil

}

// CreateJiraIssue, given the description create a Jira issue with the
// summary, accountID, issue type, project from the JiraEnv
func (j *JiraEnv) CreateJiraIssue(description string) (*jira.Issue, error) {
    fields := &jira.IssueFields{
        Reporter: &jira.User{
            AccountID: j.JiraUserAccountID,
        },
        Description: description,
        Type: jira.IssueType{
            Name: j.JiraIssueType,
        },
        Project: jira.Project{
            Key: j.JiraProject,
        },
        Summary:  j.JiraSummary, 
    }

    issue := &jira.Issue{
        Fields: fields, 
    }

    createdIssue, resp, err := j.JiraClient.createIssue(issue) 

    if err != nil {
      bodyBytes, _ := ioutil.ReadAll(resp.Body)
      return nil, fmt.Errorf("get self err: %s", bodyBytes)

    }

    return createdIssue, nil

}
