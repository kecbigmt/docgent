package application

import (
	"encoding/json"
	"net/http"

	"docgent-backend/internal/model/infrastructure"
	"docgent-backend/internal/workflow"
)

type GenerateDocumentHandler struct {
	documentationAgent infrastructure.DocumentationAgent
}

func NewGenerateDocumentHandler(agent infrastructure.DocumentationAgent) *GenerateDocumentHandler {
	return &GenerateDocumentHandler{
		documentationAgent: agent,
	}
}

func (h *GenerateDocumentHandler) Pattern() string {
	return "/api/generate"
}

type GenerateDocumentRequest struct {
	Text string `json:"text"`
}

type GenerateDocumentResponse struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *GenerateDocumentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req GenerateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	draftGenerateWorkflow := workflow.NewDraftGenerateWorkflow(workflow.DraftGenerateWorkflowParams{
		DocumentationAgent: h.documentationAgent,
	})
	draft, err := draftGenerateWorkflow.Execute(r.Context(), req.Text)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GenerateDocumentResponse{
		Title:   draft.Title,
		Content: draft.Content,
	})
}
