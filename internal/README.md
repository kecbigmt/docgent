# /internal レイヤー構成

- /domain
  - Docgent のビジネスロジック本体や重要な概念を表現するレイヤー
  - インターフェイスを用いて、標準パッケージ以外には依存しないようにする（外部ライブラリ・別の層への依存は不可）
- /application
  - システムが提供するユースケースを実装するレイヤー  
  - ユーザーの目的を達成するために、複数のドメインロジックや外部サービスをオーケストレーションする
  - 標準パッケージまたはdomain層にのみ依存可能。外部ライブラリ・domain以外の層への依存は不可
  - 外部システムとのやりとりを必要とする場合はportでインターフェイスを定義する
- /infrastructure
  - 外部サービス連携やデータ永続化などの技術的な実装を担うレイヤー
  - DB / 外部API など、技術的関心を扱うアダプタの実装を集約する
  - 種類
    - クライアント（`github`, `slack`, `google/vertexai/genai` など）
      - DBや外部APIと連携するためのクライアント
      - `application` が必要とするインターフェイスを実装する
      - domain層・application層・外部ライブラリに依存可能。ハンドラーへの依存は不可
    - ハンドラー（`handler`）
      - Slack のイベントハンドラー（`slack_events_handler.go` など）、GitHub Webhook のハンドラー（`github_events_handler.go` など）  
      - Webhook リクエストをパースし、DTOに変換して `application` のユースケースを呼び出す
      - domain層・application層・外部ライブラリに依存可能。クライアントへの依存も可

