package plugin

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

// PostBotDM posts a DM as the cloud bot user.
func (p *Plugin) PostBotDM(userID string, message string) error {
	return p.createBotPostDM(&model.Post{
		UserId:  p.BotUserID,
		Message: message,
	}, userID)
}

func (p *Plugin) createBotPostDM(post *model.Post, userID string) error {
	channel, err := p.API.GetDirectChannel(userID, p.BotUserID)
	if err != nil {
		return errors.Wrap(err, "failed to get direct channel for bot")
	}

	if channel == nil {
		return errors.Wrap(err, "failed to get direct channel for bot and user_id")
	}

	post.ChannelId = channel.Id
	_, err = p.API.CreatePost(post)
	if err != nil {
		return errors.Wrap(err, "failed to create post")
	}

	return nil
}
