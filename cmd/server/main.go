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
	"docgent-backend/internal/infrastructure/genkit"
	"docgent-backend/internal/infrastructure/github"
)

func main() {
	fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Provide(
			NewSlackAPI,
			NewGitHubAPI,
			NewGenkitConfig,
			NewHTTPServer,
			fx.Annotate(
				NewServeMux,
				fx.ParamTags(`group:"routes"`),
			),
			AsRoute(application.NewEchoHandler),
			AsRoute(application.NewHelloHandler),
			AsRoute(application.NewSlackEventHandler),
			AsSlackEventRoute(application.NewSlackReactionAddedEventConsumer),
			AsDocumentAgent(genkit.NewDocumentAgent),
			AsPoposalAgent(genkit.NewProposalAgent),
			AsGitHubBranchAPIFactory(github.NewBranchAPIFactory),
			AsGitHubPullRequestAPIFactory(github.NewPullRequestAPIFactory),
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

func AsDocumentAgent(f any, anns ...fx.Annotation) any {
	anns = append([]fx.Annotation{fx.As(new(domain.DocumentAgent))}, anns...)
	return fx.Annotate(f, anns...)
}

func AsPoposalAgent(f any, anns ...fx.Annotation) any {
	anns = append([]fx.Annotation{fx.As(new(domain.ProposalAgent))}, anns...)
	return fx.Annotate(f, anns...)
}

func AsGitHubBranchAPIFactory(f any, anns ...fx.Annotation) any {
	anns = append([]fx.Annotation{fx.As(new(application.GitHubBranchAPIFactory))}, anns...)
	return fx.Annotate(f, anns...)
}

func AsGitHubPullRequestAPIFactory(f any, anns ...fx.Annotation) any {
	anns = append([]fx.Annotation{fx.As(new(application.GitHubPullRequestAPIFactory))}, anns...)
	return fx.Annotate(f, anns...)
}
