# Docgent

Docgentは、社内のチャットをもとにドキュメントを作成・更新して、同じことを2回答える必要がないようにするためのAIエージェントです。

## 主な機能



- ドキュメントの作成・更新の提案
  - 1. ユーザーがSlackのスレッドで <img src="doc_it.png" width="20"> の絵文字をつける
  - 2. AIエージェントがスレッド内の会話をもとにして、GitHubリポジトリでPull Requestを作成し、ドキュメントの作成・更新を提案する
- フィードバックによる提案の修正
  - Pull Request上でのユーザーがフィードバックをすると、AIエージェントが提案内容を修正する
- RAGコーパスへのファイル追加
  - リポジトリのmainブランチにファイルがプッシュされると、自動的にRAGコーパスにファイルを追加される
  - Google CloudのRAG Engine APIを利用
  - 更新されたRAGコーパスは次のドキュメント作成・更新、ユーザーからの質問への回答に利用される
- ドキュメントに基づく質問回答
  - Slack上でボットにメンションして質問をすると、RAGコーパスに基づき回答を生成してくれる

## セットアップの前提条件

- Slack
  - [Slack App](https://api.slack.com/apps)が作成されていて、Signing Secret と Bot User OAuth Token を取得していること
    - Signing Secret: _Basic Information_ から取得
    - Bot User OAuth Token: _OAuth & Permissions_ から取得
  - ボットユーザーに以下のスコープの権限が付与されていること
    - app_mentions:read
    - channels:history
    - chat:write
    - reactions:read
    - reactions:write
    - users:read
    - im:history （DM で使いたいときだけ）
    - groups:history （プライベートチャンネルで使いたいときだけ）
  - ボットがワークスペースにインストールされていること
  - ワークスペースに `doc_it` の名前で絵文字が登録されていること
    - 素材: <img src="doc_it.png" width="20">
    - 画像は何でも可
- GitHub
  - [GitHub App](https://github.com/settings/apps)が作成されていて、App ID・Webhook Secret・秘密鍵が発行されていること
    - App ID: _General_ > _About_ に書いてある
    - Webhook Secret: _General_ > _Webhook_ から Webhook を有効にする
    - 秘密鍵: _General_ > _Private Keys_ から生成・ダウンロードする
  - GitHub App に以下の権限がついていること（_Permissions & events_ > _Permissions_）
    - Metadata: `Read only`
    - Contents: `Read and write`
    - Issues: `Read and write`
    - Pull Requests: `Read and write`
  - GitHub App が以下のイベントを購読していること（_Permissions & events_ > _Subscribe to events_）
    - Issue comment
    - Push
  - ドキュメント管理用のリポジトリが作成されていて、オーナー名・リポジトリ名が決まっていること
    　- 既存リポジトリでも構いませんが、最初は新規作成をおすすめします
    - オーナー名・リポジトリ名はリポジトリの URL（`https://github.com/OWNER/REPO`）から抜き出せます
  - ドキュメント管理用のリポジトリに、上記で作成した GitHub App がインストールされていて、インストール ID が発行されていること
    - GitHub App 管理画面の _General_ > _Install App_ から、自分が管理しているリポジトリにアプリをインストールできます
    - インストール完了後、対象リポジトリの _Settings_ > _Integrations_ > _GitHub Apps_ にある対象アプリの設定画面の URL にインストール ID が入っています
      - e.g. `https://github.com/apps/[アプリ名]/installations/[インストールID]`
- Google Cloud
  - プロジェクトが作成されていること
  - Vertex AI API が有効になっていること
  - 課金が有効になっていること

## インストール

### Cloud Run のデプロイ

[![Run on Google Cloud](https://storage.googleapis.com/cloudrun/button.svg)](https://console.cloud.google.com/cloudshell/editor?shellonly=true&cloudshell_image=gcr.io/cloudrun/button&cloudshell_git_repo=https%3A%2F%2Fgithub.com%2Fkecbigmt%2Fdocgent)

環境変数をたくさん設定するように案内されます。後述する環境変数の一覧を参考にしてください。

デプロイが完了したら、Cloud Run が決めてくれる URL を使って、Slack・GitHub の Webhook エンドポイントの設定を行ってください。

#### Slack App の Webhook エンドポイントを登録

[Slack App](https://api.slack.com/apps)でアプリを選択 > _Features_ > Event Subscriptions から設定する

```
https://xxxxx.a.run.app/api/slack/events
```

### GitHub App の Webhook エンドポイントを登録

[GitHub App](https://github.com/settings/apps)でアプリを選択 > _General_ > _Webhook_ から設定する

```
https://xxxxx.a.run.app/api/github/events
```

## 開発環境のセットアップ

### ソースコードの取得

```sh
git clone git@github.com:kecbigmt/docgent.git
cd docgent
```

### 環境変数の設定

以下のように環境変数をセットしてください

```sh
# Slack App設定
export SLACK_BOT_TOKEN=xoxb-123456789
export SLACK_SIGNING_SECRET=xxxxxxxxxxxx

# Slackワークスペース設定
export SLACK_WORKSPACE_ID=T0X0X0X0X

# GitHub App設定
export GITHUB_APP_PRIVATE_KEY=$(cat /path/to/private-key.pem)
export GITHUB_APP_ID=1234567
export GITHUB_WEBHOOK_SECRET=0x0x0x0x0x0x0x

# インストール先のGitHubリポジトリ
export GITHUB_OWNER=xxxxxx
export GITHUB_REPO=xxxxxx
export GITHUB_INSTALLATION_ID=123456789

# Vertex AIの設定
export VERTEXAI_PROJECT_ID=xxxxxx # Google CloudプロジェクトのID
export VERTEXAI_LOCATION=us-central1
export VERTEXAI_MODEL_NAME=gemini-2.0-flash
export VERTEXAI_RAG_CORPUS_ID=123456789123456789 # 別途作成後にセット。作成するまでは記載しない

# Google Cloudの認証情報（上記のVertex AIを利用できる権限を持っていること）
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/application_default_credentials.json
```

### サーバーを起動

```
go run ./cmd/server/*.go
```

### サーバーをインターネットに公開

注意

- ngrok が発行する URL を外部に公開しないこと
- 動作確認が終わったら ngrok を終了すること

```
ngrok http 8080
```

払い出される URL を使って、Slack App と GitHub App にエンドポイントの設定をすると、Slack や GitHub でのイベントがローカルのサーバーに届いてアプリが動作します。
