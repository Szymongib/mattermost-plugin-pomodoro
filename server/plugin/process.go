package plugin

import (
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

const (
	batchSize = 500 // TODO: adjust batch size
)

func (p *Plugin) EnqueueSessions() error {
	keys, err := p.ListAllKeys()
	if err != nil {
		return errors.Wrap(err, "failed to list keys")
	}

	for _, k := range keys {
		if strings.HasSuffix(k, activeSessionKey) {
			userID := strings.TrimSuffix(k, activeSessionKey)

			session, err := p.GetActiveSession(userID)
			if err != nil {
				return errors.Wrapf(err, "failed to get active session for user: %s", userID)
			}

			p.SessionQueue.Add(&session)
		}
	}

	return nil
}

func (p *Plugin) finalizeSession(userID string) error {
	session, err := p.GetActiveSession(userID)
	if err != nil {
		return errors.Wrapf(err, "failed to get active session for user: %s", userID)
	}

	// TODO: possibly set to previous status or one defined in config
	_, appErr := p.API.UpdateUserStatus(userID, model.STATUS_ONLINE)
	if appErr != nil {
		return errors.Wrap(err, "failed to set user status to 'online'")
	}

	err = p.SaveFinishedSession(userID, session)
	if err != nil {
		return errors.Wrap(err, "failed to save finished session")
	}

	// TODO: consider adding session ID / Name to the DM
	// TODO: might also check if the message was already sent - GetPostsSince(session finished time)
	err = p.PostBotDM(userID, "Your pomodoro session is finished!")
	if err != nil {
		p.API.LogError("FAILED TO DM!!")
		return errors.Wrap(err, "failed to create bot DM")
	}

	appErr = p.API.KVDelete(userID + activeSessionKey)
	if appErr != nil {
		return errors.Wrap(appErr, "failed to delete active session")
	}

	return nil
}
