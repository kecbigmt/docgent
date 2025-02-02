package main

import (
	"github.com/alecthomas/kong"
)

var cli struct {
	Build BuildCmd `cmd:"" help:"システムプロンプトをテンプレートから組み立て"`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("prompt-tool"),
		kong.Description("プロンプトテンプレートアセンブラ"),
		kong.UsageOnError(),
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
