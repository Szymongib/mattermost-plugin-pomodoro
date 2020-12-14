package plugin

import (
	"fmt"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
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
		return &model.CommandResponse{Text: "Need one argument - pomodoro session length"}
	}

	sessionLength, err := time.ParseDuration(params[0])
	if err != nil {
		p.API.LogWarn("Invalid session length", "error", err)
		return &model.CommandResponse{Text: fmt.Sprintf("Invalid session length: %s. Should be in format: 30m, 1h etc.", params[0])}
	}

	endTime, appErr := p.startWorkSession(args.UserId, sessionLength)
	if appErr != nil {
		p.API.LogWarn("Failed to start pomodoro session", "error", appErr)
		return &model.CommandResponse{Text: appErr.Message()}
	}

	return &model.CommandResponse{Text: fmt.Sprintf("Pomodoro session started, end at: %s", endTime.Format(dateTimeLayout))}
}

func (p *Plugin) cmdListSessions(_ *plugin.Context, args *model.CommandArgs, params []string) *model.CommandResponse {
	sessions, err := p.GetUserSessions(args.UserId)
	if err != nil {
		p.API.LogWarn("Failed to get sessions", "error", err)
		return &model.CommandResponse{Text: "Failed to list pomodoro sessions"}
	}

	message := sessionsAsMessage(sessions)

	err = p.PostBotDM(args.UserId, message)
	if err != nil {
		p.API.LogWarn("Failed send DM with sessions", "error", err)
		return &model.CommandResponse{Text: "Failed to list pomodoro sessions"}
	}

	return &model.CommandResponse{Text: "Sessions details sent in DM"}
}

func sessionsAsMessage(sessions UserSessions) string {
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
