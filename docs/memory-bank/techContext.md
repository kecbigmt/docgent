# Technical Context

## Technologies Used

### Programming Language
- **Go (Golang)**: The entire application is written in Go, leveraging its strong concurrency support, performance, and simplicity.

### Frameworks & Libraries
- **fx (Uber)**: Used for dependency injection and application lifecycle management.
- **zap (Uber)**: Structured logging library.
- **slack-go/slack**: Official Slack API client for Go.
- **google/go-github**: GitHub API client for Go.
- **bradleyfalzon/ghinstallation**: GitHub App authentication for Go.
- **Google Vertex AI**: Used for RAG (Retrieval-Augmented Generation) and LLM capabilities.

### External Services
- **Slack**: Primary user interface for triggering document generation and asking questions.
- **GitHub**: Used for storing documentation and managing the review process through pull requests.
- **Google Cloud Platform**:
  - **Vertex AI**: Provides the LLM capabilities for document generation and question answering.
  - **RAG Engine API**: Creates and manages RAG corpora for document retrieval and directly uploads files to the corpus.

### Infrastructure
- **Docker**: Application is containerized for consistent deployment.
- **Cloud Build**: Used for CI/CD pipeline.

## Development Setup

### Local Development
The application uses environment variables for configuration, with different settings for development and production environments:
- `.env.development`: Contains configuration for local development.

### API Keys & Authentication
- **Slack**: Requires a bot token and signing secret.
- **GitHub**: Uses GitHub App authentication with an app ID and private key.
- **Google Cloud**: Requires appropriate service account credentials.

## Technical Constraints

### Statelessness
- The application is designed to be stateless, with no persistent database of its own.
- All state is maintained in external systems (GitHub repositories, RAG corpora).

### API Rate Limits
- Subject to rate limits from Slack, GitHub, and Google Cloud APIs.
- Implementation includes appropriate error handling and backoff strategies.

### Security Considerations
- Webhook verification for both Slack and GitHub to ensure request authenticity.
- Secure handling of API tokens and private keys.
- Proper permission scoping for GitHub and Slack integrations.

## Dependencies

### Runtime Dependencies
- **Slack API**: For receiving events and sending messages.
- **GitHub API**: For creating branches, pull requests, and managing files.
- **Vertex AI API**: For LLM-based document generation and RAG-based question answering.

### Development Dependencies
- **Go Modules**: For dependency management.
- **Testing Libraries**: Standard Go testing package with appropriate mocks for external services.

## Architecture Decisions

### Hybrid Architecture
- The application follows a hybrid architecture that combines elements of Hexagonal, Clean, and Onion architectures.
- Domain layer contains both core business logic and some domain-specific interfaces (following DDD principles).
- Application layer defines additional ports for external service integration.
- All dependencies point inward, maintaining architectural integrity.
- This approach enables easier testing and potential replacement of external services.

### Tool-based Agent Pattern
- The AI agent operates through a tool-use pattern, where it can request specific actions (like file creation or RAG queries) through a structured XML format.
- This provides a clear separation between the agent's decision-making and the execution of those decisions.
- Tools are defined in the domain layer and implemented by handlers in the application layer.

### Event-Driven Design
- The application is primarily event-driven, responding to webhooks from Slack and GitHub.
- This enables asynchronous processing and better scalability.

### RAG for Knowledge Retrieval
- Uses Retrieval-Augmented Generation to provide context to the LLM when answering questions or generating documentation.
- This improves accuracy and relevance of responses by grounding them in existing documentation.
- Files are directly uploaded to the RAG corpus using the RAG Engine API without intermediate storage.
