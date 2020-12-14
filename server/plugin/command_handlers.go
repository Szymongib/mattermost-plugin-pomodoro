package plugin

import (
	"fmt"
	model2 "github.com/mattermost/mattermost-plugin-pomodoro/server/model"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"strconv"
	"strings"
	"time"
)

var (
	sessionsMessageFormat = `Your past sessions:
%s
`
	sessionMessageEntry = ` - On: %s, Length: %s`
	sessionLengthFormat = `%d minutes %d seconds`
	dateTimeLayout      = "2006-01-02 15:04:05"
)

type CommandHandleFunc func(c *plugin.Context, args *model.CommandArgs, parameters []string) *model.CommandResponse

func (p *Plugin) cmdStartSession(_ *plugin.Context, args *model.CommandArgs, params []string) *model.CommandResponse {
	if len(params) == 0 {
		p.API.LogDebug("Command called with insufficient parameters")
		return &model.CommandResponse{Text: "Need one argument - session length"}
	}

	sessionLength, err := strconv.Atoi(params[0]) // TODO: parse from time
	if err != nil {
		p.API.LogWarn("Invalid session length", "error", err)
		return &model.CommandResponse{Text: fmt.Sprintf("Invalid session length: %s", params[0])}
	}

	endTime, appErr := p.startWorkSession(args.UserId, time.Duration(sessionLength)*time.Second)
	if appErr != nil {
		p.API.LogWarn("Failed to start session", "error", appErr)
		return &model.CommandResponse{Text: appErr.Message()}
	}

	return &model.CommandResponse{Text: fmt.Sprintf("Session started, end at: %s", endTime.Format(dateTimeLayout))}
}

func (p *Plugin) cmdListSessions(_ *plugin.Context, args *model.CommandArgs, params []string) *model.CommandResponse {
	sessions, err := p.GetUserSessions(args.UserId)
	if err != nil {
		p.API.LogWarn("Failed to get sessions", "error", err)
		return &model.CommandResponse{Text: "Failed to list sessions"}
	}

	message := sessionsAsMessage(sessions)

	err = p.PostBotDM(args.UserId, message)
	if err != nil {
		p.API.LogWarn("Failed send DM with sessions", "error", err)
		return &model.CommandResponse{Text: "Failed to list sessions"}
	}

	return &model.CommandResponse{Text: "Sessions details sent in DM"}
}

func sessionsAsMessage(sessions model2.UserSessions) string {
	entries := make([]string, 0, len(sessions.Items))

	for _, sess := range sessions.Items {
		entries = append(
			entries,
			fmt.Sprintf(sessionMessageEntry,
				time.Unix(sess.StartTime, 0).Format(dateTimeLayout),
				fmt.Sprintf(sessionLengthFormat, sess.Length/60, sess.Length%60),
			),
		)
	}

	return fmt.Sprintf(sessionsMessageFormat, strings.Join(entries, "\n"))
}
