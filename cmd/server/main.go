package main

import (
	"context"
	"net"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"docgent-backend/internal/infrastructure/github"
	"docgent-backend/internal/infrastructure/google/vertexai/genai"
	"docgent-backend/internal/infrastructure/handler"
	"docgent-backend/internal/infrastructure/slack"
)

func main() {
	fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Provide(
			newApplicationConfigServiceFromEnv,
			newSlackAPI,
			newGitHubAPI,
			newGitHubWebhookRequestParser,
			newGenAIConfig,
			newRAGService,
			newHTTPServer,
			slack.NewServiceProvider,
			fx.Annotate(
				newServeMux,
				fx.ParamTags(`group:"routes"`),
			),
			asRoute(handler.NewHealthHandler),
			asRoute(handler.NewSlackEventHandler),
			asRoute(handler.NewGitHubWebhookHandler),
			asSlackEventRoute(handler.NewSlackReactionAddedEventConsumer),
			asSlackEventRoute(handler.NewSlackMentionEventConsumer),
			asGitHubEventRoute(handler.NewGitHubIssueCommentEventConsumer),
			asGitHubEventRoute(handler.NewGitHubPushEventConsumer),
			genai.NewChatModel,
			github.NewServiceProvider,
			zap.NewExample,
		),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}

func newHTTPServer(lc fx.Lifecycle, mux *http.ServeMux, log *zap.Logger) *http.Server {
	srv := &http.Server{Addr: ":8080", Handler: mux}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			log.Info("Starting HTTP server", zap.String("addr", srv.Addr))
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
	return srv
}

func newServeMux(routes []route) *http.ServeMux {
	mux := http.NewServeMux()
	for _, route := range routes {
		mux.Handle(route.Pattern(), route)
	}
	return mux
}

func asRoute(f any, anns ...fx.Annotation) any {
	anns = append([]fx.Annotation{fx.As(new(route)), fx.ResultTags(`group:"routes"`)}, anns...)
	return fx.Annotate(f, anns...)
}

func asSlackEventRoute(f any, anns ...fx.Annotation) any {
	anns = append([]fx.Annotation{fx.As(new(handler.SlackEventRoute)), fx.ResultTags(`group:"slack_event_routes"`)}, anns...)
	return fx.Annotate(f, anns...)
}

func asGitHubEventRoute(f any, anns ...fx.Annotation) any {
	anns = append([]fx.Annotation{fx.As(new(handler.GitHubEventRoute)), fx.ResultTags(`group:"github_event_routes"`)}, anns...)
	return fx.Annotate(f, anns...)
}
