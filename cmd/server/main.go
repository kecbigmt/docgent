package main

import (
	"context"
	"net"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"docgent-backend/internal/application"
	"docgent-backend/internal/domain"
	"docgent-backend/internal/infrastructure/github"
	"docgent-backend/internal/infrastructure/google/vertexai/genai"
	"docgent-backend/internal/infrastructure/slack"
)

func main() {
	fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Provide(
			NewSlackAPI,
			NewGitHubAPI,
			NewGitHubWebhookRequestParser,
			NewVertexAIConfig,
			NewHTTPServer,
			slack.NewServiceProvider,
			fx.Annotate(
				NewServeMux,
				fx.ParamTags(`group:"routes"`),
			),
			AsRoute(application.NewEchoHandler),
			AsRoute(application.NewHelloHandler),
			AsRoute(application.NewSlackEventHandler),
			AsRoute(application.NewGitHubWebhookHandler),
			AsSlackEventRoute(application.NewSlackReactionAddedEventConsumer),
			AsGitHubEventRoute(application.NewGitHubIssueCommentEventConsumer),
			AsChatModel(genai.NewChatModel),
			AsGitHubServiceProvider(github.NewServiceProvider),
			zap.NewExample,
		),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}

func NewHTTPServer(lc fx.Lifecycle, mux *http.ServeMux, log *zap.Logger) *http.Server {
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

func NewServeMux(routes []application.Route) *http.ServeMux {
	mux := http.NewServeMux()
	for _, route := range routes {
		mux.Handle(route.Pattern(), route)
	}
	return mux
}

func AsRoute(f any, anns ...fx.Annotation) any {
	anns = append([]fx.Annotation{fx.As(new(application.Route)), fx.ResultTags(`group:"routes"`)}, anns...)
	return fx.Annotate(f, anns...)
}

func AsSlackEventRoute(f any, anns ...fx.Annotation) any {
	anns = append([]fx.Annotation{fx.As(new(application.SlackEventRoute)), fx.ResultTags(`group:"slack_event_routes"`)}, anns...)
	return fx.Annotate(f, anns...)
}

func AsGitHubEventRoute(f any, anns ...fx.Annotation) any {
	anns = append([]fx.Annotation{fx.As(new(application.GitHubEventRoute)), fx.ResultTags(`group:"github_event_routes"`)}, anns...)
	return fx.Annotate(f, anns...)
}

func AsChatModel(f any, anns ...fx.Annotation) any {
	anns = append([]fx.Annotation{fx.As(new(domain.ChatModel))}, anns...)
	return fx.Annotate(f, anns...)
}

func AsGitHubServiceProvider(f any, anns ...fx.Annotation) any {
	anns = append([]fx.Annotation{fx.As(new(application.GitHubServiceProvider))}, anns...)
	return fx.Annotate(f, anns...)
}
