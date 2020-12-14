package plugin

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

const (
	minSessionLength = 10 * time.Second
	maxSessionLength = 24 * time.Hour
)

var (
	invalidSessionLen = fmt.Sprintf("Failed to start session: session length must be between %s and %s", minSessionLength, maxSessionLength)
	sessionInProgress = "Failed to start session: session for user already in progress"
)

func (p *Plugin) startWorkSession(userID string, sessionLength time.Duration) (time.Time, *PluginError) {
	if sessionLength < minSessionLength || sessionLength > maxSessionLength {
		return time.Time{}, Err(fmt.Errorf(invalidSessionLen), invalidSessionLen)
	}

	hasSession, err := p.HasActiveSession(userID)
	if err != nil {
		return time.Time{}, InternalErr(errors.Wrap(err, "failed to get active session for user"))
	}

	if hasSession {
		return time.Time{}, Err(fmt.Errorf(sessionInProgress), sessionInProgress)
	}

	// TODO: potentially set to preferred status defined in config
	_, appErr := p.API.UpdateUserStatus(userID, model.STATUS_DND)
	if appErr != nil {
		return time.Time{}, InternalErr(errors.Wrap(appErr, "failed to set user status to do not disturb"))
	}

	startTime := time.Now()
	session := Session{
		SessionID: model.NewId(),
		UserID:    userID,
		StartTime: startTime.Unix(),
		Length:    int64(sessionLength.Seconds()),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return time.Time{}, InternalErr(errors.Wrap(err, "failed to marshal session data"))
	}

	appErr = p.API.KVSet(userID+activeSessionKey, data)
	if appErr != nil {
		return time.Time{}, InternalErr(errors.Wrap(appErr, "failed to save active session"))
	}

	p.SessionQueue.Add(&session)

	endTime := startTime.Add(sessionLength)

	return endTime, nil
}
