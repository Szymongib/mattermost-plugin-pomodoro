package plugin

import (
	"encoding/json"
	"fmt"

	model2 "github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

const (
	userSessionsKey  = "_usersessions"
	activeSessionKey = "_activesession"

	// TODO: session lock as separate key with expiry
)

func (p *Plugin) SaveFinishedSession(userID string, session Session) error {
	userSessions, err := p.GetUserSessions(userID)
	if err != nil {
		return errors.Wrap(err, "failed to list sessions")
	}

	if sessionExists(userSessions, session.SessionID) {
		return nil
	}

	userSessions.Items = append(userSessions.Items, session)

	data, jsonErr := json.Marshal(&userSessions)
	if jsonErr != nil {
		return errors.Wrap(err, "failed to marshal user sessions")
	}

	err = p.API.KVSet(userID+userSessionsKey, data)
	if err != nil {
		return errors.Wrap(err, "failed to store user sessions")
	}

	return nil
}

func sessionExists(sessions UserSessions, id string) bool {
	for _, s := range sessions.Items {
		if s.SessionID == id {
			return true
		}
	}
	return false
}

func (p *Plugin) GetUserSessions(userID string) (UserSessions, error) {
	userSessionsRaw, appErr := p.API.KVGet(userID + userSessionsKey)
	if appErr != nil {
		return UserSessions{}, errors.Wrap(appErr, "failed to get user sessions")
	}

	// No session - return empty list
	if userSessionsRaw == nil {
		return UserSessions{Items: []Session{}}, nil
	}

	var userSessions UserSessions
	err := json.Unmarshal(userSessionsRaw, &userSessions)
	if err != nil {
		return UserSessions{}, errors.Wrap(err, "failed to unmarshal user sessions")
	}

	return userSessions, nil
}

func (p *Plugin) GetActiveSession(userID string) (Session, error) {
	sessionRaw, appErr := p.getActiveSession(userID)
	if appErr != nil {
		return Session{}, errors.Wrapf(appErr, "failed to query active session")
	}

	if sessionRaw == nil {
		return Session{}, fmt.Errorf("session not found for the user")
	}

	var session Session
	err := json.Unmarshal(sessionRaw, &session)
	if err != nil {
		return Session{}, errors.Wrapf(err, "failed to unmarshall active session")
	}

	return session, nil
}

func (p *Plugin) HasActiveSession(userID string) (bool, error) {
	data, err := p.getActiveSession(userID)
	if err != nil {
		return false, errors.Wrap(err, "failed to get session for user")
	}

	// TODO: here check if session is already finished

	return data != nil, nil
}

func (p *Plugin) getActiveSession(userID string) ([]byte, *model2.AppError) {
	return p.API.KVGet(userID + activeSessionKey)
}

func (p *Plugin) ListAllKeys() ([]string, error) {
	var keys []string
	offset := 0

	for {
		keyBatch, err := p.API.KVList(offset, batchSize)
		if err != nil {
			return nil, errors.Wrap(err, "failed to list keys")
		}
		keys = append(keys, keyBatch...)
		if len(keyBatch) < batchSize {
			break
		}
		offset += 1
	}

	return keys, nil
}
