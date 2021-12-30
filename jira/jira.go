package jira

import (
    "context"
    "io/ioutil"
	"github.com/andygrunwald/go-jira"
    "fmt"
)

type JiraEnv struct {
    JiraContext context.Context
	JiraClient    *jira.Client
    JiraUrl string
    JiraProject string
    JiraSummary string
    JiraIssueType string
    JiraUserAccountID string
}

type JiraClientWrapper interface {
    getSelf() (*jira.User, *jira.Response, error)
    createIssue(*jira.Issue) (*jira.Issue, *jira.Response, error)


}

func NewEnv(jiraUrl string, username string, password string, jiraProject string, jiraSummary string, jiraIssueType string) (*JiraEnv, error) {
    tp := jira.BasicAuthTransport{
        Username: username,
        Password: password,
    }

    jiraClient, _ := jira.NewClient(tp.Client(), jiraUrl)

    env := &JiraEnv{
        JiraContext: context.Background(),
        JiraClient: jiraClient,
        JiraUrl: jiraUrl,
        JiraProject: jiraProject,
        JiraSummary: jiraSummary,
        JiraIssueType: jiraIssueType,
    }

    jiraUser, resp, err := env.getSelf()

    if err != nil {
        bodyBytes, _ := ioutil.ReadAll(resp.Body)
        return nil, fmt.Errorf("get self err: %s", bodyBytes)

    }

    return env.setAccountID(jiraUser.AccountID), nil

}

func (j *JiraEnv) createIssue(issue *jira.Issue) (*jira.Issue, *jira.Response, error) {
    return j.JiraClient.Issue.CreateWithContext(j.JiraContext, issue)

}

func (j *JiraEnv) getSelf() (*jira.User, *jira.Response, error) {
    return j.JiraClient.User.GetSelf()

}

func (j *JiraEnv) setAccountID(accountID string) *JiraEnv {
    j.JiraUserAccountID = accountID
    return j
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


    createdIssue, resp, err := j.createIssue(issue) 

    if err != nil {
      bodyBytes, _ := ioutil.ReadAll(resp.Body)
      return nil, fmt.Errorf("get self err: %s", bodyBytes)

    }

    return createdIssue, nil

}
