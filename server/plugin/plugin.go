package plugin

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

// TODO: tests

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	BotUserID string

	CommandHandlers map[string]CommandHandleFunc

	SessionQueue *SessionQueue
}

// TODO: add bot icon
func (p *Plugin) OnActivate() error {
	config := p.getConfiguration()

	if err := config.IsValid(); err != nil {
		return errors.Wrap(err, "invalid config")
	}

	botID, err := p.Helpers.EnsureBot(&model.Bot{
		Username:    "pomodoro",
		DisplayName: "Pomodoro",
		Description: "Created by the Pomodoro plugin.",
	})
	if err != nil {
		return errors.Wrap(err, "failed to ensure pomodoro bot")
	}
	p.BotUserID = botID

	p.SessionQueue = p.NewWorkQueue(7)

	// TODO: handle multiple plugin instances
	err = p.EnqueueSessions()
	if err != nil {
		p.API.LogError("Failed to enqueue sessions", "error", err)
		return errors.Wrap(err, "failed to enqueue active sessions")
	}

	return nil
}

// NewPlugin returns an instance of a Plugin.
func NewPlugin() *Plugin {
	p := &Plugin{}

	p.CommandHandlers = map[string]CommandHandleFunc{
		"start": p.cmdStartSession,
		"list":  p.cmdListSessions,
	}

	return p
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world!")
}
