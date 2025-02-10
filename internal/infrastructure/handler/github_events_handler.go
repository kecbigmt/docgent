package handler

import (
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"docgent/internal/infrastructure/github"
)

type GitHubWebhookHandlerParams struct {
	fx.In

	Logger                     *zap.Logger
	EventRoutes                []GitHubEventRoute `group:"github_event_routes"`
	GitHubWebhookRequestParser *github.WebhookRequestParser
}

type GitHubWebhookHandler struct {
	logger                     *zap.Logger
	eventRoutes                []GitHubEventRoute
	githubWebhookRequestParser *github.WebhookRequestParser
}

func NewGitHubWebhookHandler(params GitHubWebhookHandlerParams) *GitHubWebhookHandler {
	return &GitHubWebhookHandler{
		logger:                     params.Logger,
		eventRoutes:                params.EventRoutes,
		githubWebhookRequestParser: params.GitHubWebhookRequestParser,
	}
}

func (h *GitHubWebhookHandler) Pattern() string {
	return "/api/github/events"
}

func (h *GitHubWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ev, err := h.githubWebhookRequestParser.ParseRequest(r)
	if err != nil {
		h.logger.Warn("Failed to parse request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	routeMap := map[string]GitHubEventRoute{}
	for _, route := range h.eventRoutes {
		routeMap[route.EventType()] = route
	}

	for _, route := range h.eventRoutes {
		if route.EventType() == ev.EventType() {
			go route.ConsumeEvent(ev.InnerEvent())
			break
		}
	}

	// Githubイベントには即座に200 OKを返す
	w.WriteHeader(http.StatusOK)
}
