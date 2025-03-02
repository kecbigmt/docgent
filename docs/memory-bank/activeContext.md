# Active Context

## Current Work Focus

The team is currently implementing the user story "Enable verification of citation sources in question responses." This involves:

1. Changing the RAG corpus file registration to use GitHub permalinks as displayName instead of file paths
2. Migrating existing corpus entries to the new permalink-based displayName format
3. Addressing architectural concerns in the citation implementation

## Recent Changes

Recent development has focused on enhancing the citation system, improving architecture, and strengthening test coverage:

1. **Enhanced Citation System**: The system has been updated to show the sources of information used when answering questions. This was implemented by:
   - Modifying the `AttemptComplete` tool structure to support multiple messages with source references
   - Enhancing the `AttemptCompleteHandler` to format responses with proper citations
   - Implementing parsing logic to generate user-friendly messages in Slack with source references

2. **Presenter Pattern Implementation**: Implemented the Presenter Pattern to abstract presentation logic:
   - Created a `ResponseFormatter` interface in the application/port package
   - Implemented concrete formatters in the infrastructure/slack and infrastructure/github packages
   - Updated the `AttemptCompleteHandler` to use the formatter
   - Updated the `ConversationUsecase` and `ProposalRefineUsecase` to accept and use the formatter
   - Updated the `ProposalGenerateUsecase` to use the ResponseFormatter
   - Updated the service providers to create formatters

3. **Improved Test Coverage**: Enhanced unit tests for the ResponseFormatter integration:
   - Added argument verification in `AttemptCompleteHandler_Handle` tests to ensure the correct `AttemptComplete` object is passed to the formatter
   - Updated `ConversationUsecase`, `ProposalRefineUsecase`, and `ProposalGenerateUsecase` tests to verify the arguments passed to the formatter
   - Fixed compatibility issues between expected and actual objects in tests by using `nil` for empty source arrays

4. **Autonomous Agent Behavior**: Instead of following a fixed workflow for RAG-based question answering, the agent now has more autonomy to determine its behavior, similar to the document creation process.

5. **Conversation Management**: Replaced the dedicated `QuestionAnswerUsecase` with a more general `ConversationUsecase` that can handle various types of interactions and maintain conversation history.

6. **Enhanced History Retrieval**: Updated both GitHub and Slack conversation services to return structured conversation history and include user mention checks.

## Next Steps

1. ✅ **Implement GitHub Permalink Integration**: The RAG corpus file registration now uses GitHub permalinks as displayName instead of file paths.

2. ✅ **Migrate Existing RAG Corpus**: A migration tool has been created to remove all files currently registered in the RAG corpus and re-register them with the new permalink-based displayName format.

3. ✅ **Citation UI Improvements**: Further refine how sources are presented to users in Slack messages with clickable links to the original documents.

## Active Decisions and Considerations

1. **GitHub Permalinks for Citations**: The team has decided to use GitHub permalinks as displayName in the RAG corpus instead of file paths. This approach is preferred over constructing URIs from page paths because:
   - It's simpler and more direct
   - It ensures the links point to the exact version of the document used as a source
   - It avoids issues with branch content changing over time

2. **Clean Architecture Enforcement**: The team has identified that the `AttemptCompleteHandler` currently contains Slack-specific message formatting logic, which violates clean architecture principles. The decision is to move this platform-specific code to the infrastructure layer.

3. **Source Citation**: When answering questions, the system now explicitly cites the sources of information, improving transparency and trustworthiness.

4. **Agent Autonomy**: The AI agent is given more flexibility to determine its own workflow rather than following a rigid process, allowing for more natural and effective interactions.

5. **Unified Conversation Management**: Moving to a more general conversation management approach rather than treating question answering as a separate workflow.
