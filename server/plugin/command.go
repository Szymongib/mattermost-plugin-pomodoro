package plugin

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	spaceRegExp := regexp.MustCompile(`\s+`)
	trimmedArgs := spaceRegExp.ReplaceAllString(strings.TrimSpace(args.Command), " ")
	stringArgs := strings.Split(trimmedArgs, " ")
	lengthOfArgs := len(stringArgs)

	if lengthOfArgs == 1 {
		return &model.CommandResponse{Text: "Subcommand not provided"}, nil
	}

	parameters := []string{}

	if lengthOfArgs > 2 {
		parameters = stringArgs[2:]
	}
	action := stringArgs[1]

	if f, ok := p.CommandHandlers[action]; ok {
		return f(c, args, parameters), nil
	}

	p.postCommandResponse(args, fmt.Sprintf("Unknown action %v", action))
	return &model.CommandResponse{}, nil

}

func (p *Plugin) getCommand(config *configuration) (*model.Command, error) {

	return &model.Command{
		Trigger:          "pomodoro",
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands: start, list",
		AutoCompleteHint: "[command]",
		AutocompleteData: getAutocompleteData(config),
	}, nil
}

func (p *Plugin) postCommandResponse(args *model.CommandArgs, text string) {
	post := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: args.ChannelId,
		RootId:    args.RootId,
		Message:   text,
	}
	_ = p.API.SendEphemeralPost(args.UserId, post)
}

func getAutocompleteData(config *configuration) *model.AutocompleteData {
	pomodoro := model.NewAutocompleteData("pomodoro", "[command]", "Available commands: start, list")

	// TODO: support time passes as: 600s / 10m / 1h etc.
	start := model.NewAutocompleteData("start", "[session_time]", "Start Pomodoro session with specified length")
	pomodoro.AddCommand(start)

	list := model.NewAutocompleteData("list", "", "Lists your past sessions as DM from Bot user")
	pomodoro.AddCommand(list)

	return pomodoro
}
