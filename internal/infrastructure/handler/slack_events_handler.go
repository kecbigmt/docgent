package handler

import (
	"encoding/json"
	"net/http"

	"github.com/slack-go/slack/slackevents"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"docgent-backend/internal/infrastructure/slack"
)

type SlackEventHandlerParams struct {
	fx.In

	Logger                   *zap.Logger
	EventRoutes              []SlackEventRoute `group:"slack_event_routes"`
	SlackAPI                 *slack.API
	ApplicationConfigService ApplicationConfigService
}

type SlackEventHandler struct {
	log                      *zap.Logger
	eventRoutes              []SlackEventRoute
	slackAPI                 *slack.API
	applicationConfigService ApplicationConfigService
}

func NewSlackEventHandler(params SlackEventHandlerParams) *SlackEventHandler {
	return &SlackEventHandler{
		log:                      params.Logger,
		eventRoutes:              params.EventRoutes,
		slackAPI:                 params.SlackAPI,
		applicationConfigService: params.ApplicationConfigService,
	}
}

func (h *SlackEventHandler) Pattern() string {
	return "/api/slack/events"
}

func (h *SlackEventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestParser := slack.NewWebhookRequestParser(h.slackAPI, h.log)

	event, err := requestParser.ParseRequest(r)
	if err != nil {
		if err == slack.ErrUnauthorizedRequest {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	workspace, err := h.applicationConfigService.GetWorkspaceBySlackWorkspaceID(event.TeamID)
	if err != nil {
		if err == ErrWorkspaceNotFound {
			h.log.Warn("Unknown Slack workspace ID", zap.String("slack_workspace_id", event.TeamID))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.log.Error("Failed to get workspace", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// URLの検証チャレンジに応答
	if event.Type == slackevents.URLVerification {
		ev, ok := event.Data.(*slackevents.ChallengeResponse)
		if !ok {
			h.log.Error("Failed to convert event data to ChallengeResponse")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"challenge": ev.Challenge,
		})
		return
	}

	// イベントの処理
	if event.Type == slackevents.CallbackEvent {
		innerEvent := event.InnerEvent

		eventJSON, err := json.Marshal(innerEvent)
		if err != nil {
			h.log.Warn("Failed to marshal inner event", zap.Error(err))
			h.log.Info("Received event", zap.Any("event", innerEvent))
		} else {
			h.log.Info("Received event", zap.String("event", string(eventJSON)))
		}

		for _, route := range h.eventRoutes {
			if innerEvent.Type == route.EventType() {
				go route.ConsumeEvent(innerEvent, workspace)
				break
			}
		}
	}

	// Slackイベントには即座に200 OKを返す
	w.WriteHeader(http.StatusOK)
}
