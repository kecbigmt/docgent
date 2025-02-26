package slack

import (
	"docgent/internal/application/port"
	"docgent/internal/domain/data"
	"fmt"
	"strings"

	"github.com/slack-go/slack"
)

type ConversationService struct {
	slackAPI   *API
	ref        *ConversationRef
	fromUserID string
}

func NewConversationService(slackAPI *API, ref *ConversationRef, fromUserID string) port.ConversationService {
	return &ConversationService{
		slackAPI:   slackAPI,
		ref:        ref,
		fromUserID: fromUserID,
	}
}

func (s *ConversationService) Reply(input string, withMention bool) error {
	slackClient := s.slackAPI.GetClient()

	message := input
	if withMention && s.fromUserID != "" {
		message = fmt.Sprintf("<@%s>\n%s", s.fromUserID, input)
	}

	_, _, err := slackClient.PostMessage(s.ref.ChannelID(), slack.MsgOptionText(message, false), slack.MsgOptionTS(s.ref.ThreadTimestamp()))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	return nil
}

func (s *ConversationService) URI() *data.URI {
	return s.ref.ToURI()
}

func (s *ConversationService) GetHistory() (port.ConversationHistory, error) {
	client := s.slackAPI.GetClient()

	messages, _, _, err := client.GetConversationReplies(&slack.GetConversationRepliesParameters{
		ChannelID: s.ref.ChannelID(),
		Timestamp: s.ref.ThreadTimestamp(),
	})
	if err != nil {
		return port.ConversationHistory{}, fmt.Errorf("failed to get thread messages: %w", err)
	}

	authTest, err := client.AuthTest()
	if err != nil {
		return port.ConversationHistory{}, fmt.Errorf("failed to auth test: %w", err)
	}
	currentUserID := authTest.UserID

	conversationMessages := make([]port.ConversationMessage, 0, len(messages))
	for _, message := range messages {
		conversationMessages = append(conversationMessages, port.ConversationMessage{
			Author:       message.User,
			Content:      message.Text,
			YouMentioned: strings.Contains(message.Text, fmt.Sprintf("@%s", currentUserID)),
			IsYou:        message.User == currentUserID,
		})
	}

	return port.ConversationHistory{
		URI:      s.ref.ToURI(),
		Messages: conversationMessages,
	}, nil
}

func (s *ConversationService) MarkEyes() error {
	slackClient := s.slackAPI.GetClient()
	err := slackClient.AddReaction("eyes", slack.ItemRef{
		Channel:   s.ref.ChannelID(),
		Timestamp: s.ref.SourceMessageTimestamp(),
	})
	if err != nil {
		return fmt.Errorf("failed to add eyes reaction: %w", err)
	}
	return nil
}

func (s *ConversationService) RemoveEyes() error {
	slackClient := s.slackAPI.GetClient()
	err := slackClient.RemoveReaction("eyes", slack.ItemRef{
		Channel:   s.ref.ChannelID(),
		Timestamp: s.ref.SourceMessageTimestamp(),
	})
	if err != nil {
		return fmt.Errorf("failed to remove eyes reaction: %w", err)
	}
	return nil
}
