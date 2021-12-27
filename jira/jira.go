package jira

import (
    "context"
    "io/ioutil"
	"github.com/andygrunwald/go-jira"
    "fmt"
)

type JiraEnv struct {
    Context context.Context
	JiraClient    *jira.Client
    JiraUrl string
    JiraProject string
    JiraSummary string
    JiraIssueType string
    JiraUserAccountID string
}

func NewEnv(jiraUrl string, username string, password string, jiraProject string, jiraSummary string, jiraIssueType string) (*JiraEnv, error) {
	tp := jira.BasicAuthTransport{
		Username: username,
		Password: password,
	}

	jiraClient, _ := jira.NewClient(tp.Client(), jiraUrl)

    jiraUser, resp, err := jiraClient.User.GetSelf()

    if err != nil {
      bodyBytes, _ := ioutil.ReadAll(resp.Body)
      return nil, fmt.Errorf("get self err: %s", bodyBytes)

    }

	return &JiraEnv{
		JiraClient:    jiraClient,
        JiraUrl: jiraUrl,
        JiraProject: jiraProject,
        JiraSummary: jiraSummary,
        JiraIssueType: jiraIssueType,
        JiraUserAccountID: jiraUser.AccountID,
        
	}, nil

}

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


    createdIssue, resp, err := j.JiraClient.Issue.CreateWithContext(context.Background(), issue)

    if err != nil {
      bodyBytes, _ := ioutil.ReadAll(resp.Body)
      return nil, fmt.Errorf("get self err: %s", bodyBytes)

    }

    return createdIssue, nil

}
