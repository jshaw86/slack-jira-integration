package slack

import (
    "fmt"
    "net/http"
    "io/ioutil"
    "bytes"

	"encoding/json"

	"github.com/slack-go/slack"

	"github.com/slack-go/slack/slackevents"
)


func ValidateSlackRequest(slackSigningSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			bodyBytes, _ := ioutil.ReadAll(req.Body)
			req.Body.Close() //  must close
			req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

			sv, err := slack.NewSecretsVerifier(req.Header, slackSigningSecret)
			if err != nil {
                fmt.Println(fmt.Sprintf("auth err: %+v", err))
				resp.WriteHeader(http.StatusBadRequest)
				return
			}
			if _, err := sv.Write(bodyBytes); err != nil {
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}
			if err := sv.Ensure(); err != nil {
				resp.WriteHeader(http.StatusUnauthorized)
				return
			}
			eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(bodyBytes), slackevents.OptionNoVerifyToken())
			if err != nil {
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}

			if eventsAPIEvent.Type == slackevents.URLVerification {
				var r *slackevents.ChallengeResponse
				err := json.Unmarshal([]byte(bodyBytes), &r)
				if err != nil {
					resp.WriteHeader(http.StatusInternalServerError)
					return
				}
				resp.Header().Set("Content-Type", "text")
				resp.Write([]byte(r.Challenge))
				return
			}

			next.ServeHTTP(resp, req)

		})
	}
}
