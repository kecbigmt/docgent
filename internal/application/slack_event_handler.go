package application

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/slack-go/slack/slackevents"
	"go.uber.org/fx"
	"go.uber.org/zap"

	appslack "docgent-backend/internal/application/slack"
)

type SlackEventHandlerParams struct {
	fx.In

	Logger                   *zap.Logger
	EventRoutes              []SlackEventRoute `group:"slack_event_routes"`
	SlackAPI                 appslack.API
	ApplicationConfigService ApplicationConfigService
}

type SlackEventHandler struct {
	log                      *zap.Logger
	eventRoutes              []SlackEventRoute
	slackAPI                 appslack.API
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
	// Slackからのリクエストを検証
	sv, err := h.slackAPI.NewSecretsVerifier(r.Header)
	if err != nil {
		h.log.Warn("Failed to create secrets verifier", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bodyReader := io.TeeReader(r.Body, &sv)
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		h.log.Warn("Failed to read request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := sv.Ensure(); err != nil {
		h.log.Warn("Failed to verify request", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	event, err := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
	if err != nil {
		h.log.Warn("Failed to parse event", zap.Error(err))
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
		var res *slackevents.ChallengeResponse
		if err := json.Unmarshal(body, &res); err != nil {
			h.log.Error("Failed to unmarshal challenge", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if res == nil {
			h.log.Error("Challenge response is nil")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"challenge": res.Challenge,
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
				return
			}
		}
	}

	// Slackイベントには即座に200 OKを返す
	w.WriteHeader(http.StatusOK)
}
