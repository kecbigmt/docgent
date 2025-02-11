# Docgent

Docgent は、社内のチャットをもとにドキュメントを作成・更新して、同じことを 2 回答える必要がないようにするための AI エージェントです。

紹介記事：https://zenn.dev/kecy/articles/84509c63e76218

## 主な機能

1. SlackのスレッドをもとにしてGitHub上でドキュメントを作成
2. GitHubのPull Requestでのコメントに基づいて編集内容を改善
3. mainブランチに反映されると、AIエージェントの知識として蓄積
4. Slackでの質問に対して、蓄積した知識をもとに回答

## デモ動画

[!['YouTube thumbnail'](https://img.youtube.com/vi/L7dzehHun18/maxres1.jpg)](https://www.youtube.com/watch?v=L7dzehHun18a "Demo video")

## セットアップの前提条件

- Slack
  - [Slack App](https://api.slack.com/apps)が作成されていること
  - ボットユーザーに以下のスコープの権限が付与されていること（_Features_ > _OAuth & Permissions_ で設定できます）
    - `app_mentions:read`
    - `channels:history`
    - `chat:write`
    - `reactions:read`
    - `reactions:write`
    - `users:read`
    - `im:history`（DM で使いたいときだけ）
    - `groups:history`（プライベートチャンネルで使いたいときだけ）
  - アプリがワークスペースにインストールされていること
    -  _OAuth & Permissions_ で権限を選択した後に `Install to [ワークスペース名]` というボタンを押すとインストールできます
  - アプリのEvent Subscriptionsが有効になっていること（_Features_ > _Event Subscriptions_）
  - アプリが以下のイベントを購読するよう設定されていること（_Event Subscriptions_ > _Subscribe to bot events_）
    - `app_mention`
    - `reaction_added`
  - ワークスペースに `doc_it` の名前で絵文字が登録されていること
    - サンプル素材: <img src="doc_it.png" width="20">
    - 画像は何でも可
- GitHub
  - [GitHub App](https://github.com/settings/apps)が作成されていること
  - アプリでWebhookが有効になっていること
    - _General_ > _Webhook_ から 有効にできます
  - アプリに以下の権限がついていること（_Permissions & events_ > _Permissions_）
    - Metadata: `Read only`
    - Contents: `Read and write`
    - Issues: `Read and write`
    - Pull Requests: `Read and write`
  - アプリが以下のイベントを購読していること（_Permissions & events_ > _Subscribe to events_）
    - Issue comment
    - Push
  - ドキュメント管理用のリポジトリが作成されていること
    - 既存リポジトリでも動作しますが、試しに使ってみる場合は新規作成をおすすめします
  - ドキュメント管理用のリポジトリに、上記で作成した GitHub App がインストールされていて、インストール ID が発行されていること
    - GitHub App 管理画面の _Install App_ から、自分が管理しているリポジトリにアプリをインストールできます
- Google Cloud
  - プロジェクトが作成されていること
  - Vertex AI API が有効になっていること
  - 課金が有効になっていること

## インストール

### Cloud Run のデプロイ

[![Run on Google Cloud](https://storage.googleapis.com/cloudrun/button.svg)](https://console.cloud.google.com/cloudshell/editor?shellonly=true&cloudshell_image=gcr.io/cloudrun/button&cloudshell_git_repo=https%3A%2F%2Fgithub.com%2Fkecbigmt%2Fdocgent)

環境変数をたくさん設定するように案内されます。以下を参考に設定してください。

変数名 | 説明
---|---
`SLACK_WORKSPACE_ID` | SlackのワークスペースID。[こちら](https://slack.com/intl/ja-jp/help/articles/221769328)を参考にして確認してください
`GITHUB_APP_ID` | GitHubのApp ID。[GitHub App](https://github.com/settings/apps) で対象アプリ選択 > _General_ > _About_ に書いてあります
`GITHUB_OWNER` | GitHubリポジトリのオーナー名。リポジトリの URL（`https://github.com/OWNER/REPO`）から抜き出せます
`GITHUB_REPO` | GitHubリポジトリの名前。同上
`GITHUB_DEFAULT_BRANCH` | GitHubリポジトリのデフォルトブランチ名。デフォルト値は `main`
`GITHUB_INSTALLATION_ID` | GitHubb AppのインストールID。対象リポジトリにインストール完了後、リポジトリの _Settings_ > _Integrations_ > _GitHub Apps_ にある対象アプリの設定画面の URL にインストール ID が入っています<br>e.g.  `https://github.com/apps/[アプリ名]/installations/[インストールID]`
`VERTEXAI_PROJECT_ID` | Vertex AIを利用できるGoogle CloudプロジェクトのID。Cloud Runと同じプロジェクトにするのを推奨します
`VERTEXAI_LOCATION` | Vertex AIを利用するリージョン名。デフォルト値は `us-central1`
`VERTEXAI_MODEL_NAME` | エージェント制御や回答生成のためのGeminiモデル名。デフォルト値は `gemini-2.0-flash`
`VERTEXAI_RAG_CORPUS_ID` | RAGコーパスのID。作成方法は後述。後回しにする場合は `0` をセットしてください（RAG 機能がオフになります）

以下の機密情報は自動で環境変数として設定されないので、初回デプロイ後に Cloud Run のコンソールから シークレット として登録してください（_新しいリビジョンの編集とデプロイ_ > _コンテナの編集_ > _変数とシークレット_）。

変数名 |  説明
---|---
`SLACK_SIGNING_SECRET` | Slack AppのSigning Secret。[Slack App](https://api.slack.com/apps)で対象アプリ選択 > _Basic Information_ から取得できます
`SLACK_BOT_TOKEN` | Slack AppのBot User OAuth Token。_OAuth & Permissions_ から取得できます
`GITHUB_WEBHOOK_SECRET` | GitHubのWebhookシークレット。[GitHub App](https://github.com/settings/apps) で対象アプリ選択 > _General_ > _Webhook_ から取得できます
`GITHUB_APP_PRIVATE_KEY` | GitHub Appからのアクセストークンリクエストに署名するための秘密鍵。_General_ > _Private Keys_ から生成・ダウンロードできます

登録後は「デプロイ」ボタンを押して再デプロイしてください。

デプロイが完了したら、Cloud Run が決めてくれる URL を使って、Slack・GitHub の Webhook エンドポイントの設定を行ってください。

#### Slack App の Webhook エンドポイントを登録

[Slack App](https://api.slack.com/apps)でアプリを選択 > _Features_ > _Event Subscriptions_ から設定する

```
https://xxxxx.a.run.app/api/slack/events
```

### GitHub App の Webhook エンドポイントを登録

[GitHub App](https://github.com/settings/apps)でアプリを選択 > _General_ > _Webhook_ から設定する

```
https://xxxxx.a.run.app/api/github/events
```

---

以上でインストール完了です。Slackワークスペースを開き、任意のスレッドで :doc_it: リアクションをつけて、応答が返ってくればOKです（ドキュメントのファイルとPull Requestが作成されます）。

## 開発環境のセットアップ

※事前に Go v1.23.4 以降をインストールしてください。

### ソースコードの取得

```sh
git clone git@github.com:kecbigmt/docgent.git
cd docgent
```

### 環境変数の設定

以下のように環境変数をセットしてください。

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

## RAG コーパスの作成

git clone した後に、以下のコマンドで CLI ツールを利用してください。

```
go run cmd/ragtool/main.go corpus create \
--project-id <Google CloudプロジェクトID> \
--location <Googe Cloudのリージョン名> \
--display-name <コーパスの表示名>
```

作成できたら、コーパスの一覧を取得して ID を確認します。

```
go run cmd/ragtool/main.go corpus list \
--project-id <Google CloudプロジェクトID> \
--location <Googe Cloudのリージョン名>
```

name の最後に入っている数字の文字列がコーパスの ID です。

```
{
  "name": "projects/xxxxxx/locations/us-central1/ragCorpora/123456789123456789",
  "displayName": "xxxxxxxx",
  "createTime": "2025-02-10T04:35:21.261097Z",
  "updateTime": "2025-02-10T04:35:21.261097Z",
  "corpusStatus": {
    "state": "ACTIVE"
  }
}
```

Cloud Run 上で動かしている場合は、コンソールから環境変数をセットして再デプロイしてください。
