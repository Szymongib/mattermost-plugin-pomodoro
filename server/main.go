package main

import (
	plugin2 "github.com/mattermost/mattermost-plugin-pomodoro/server/plugin"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func main() {
	plugin.ClientMain(plugin2.NewPlugin())

	// TODO: on start queue all active sessions
}
