package tooluse

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// Parse parses XML string and returns Union
func Parse(xmlStr string) (Union, error) {
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))

	// 最初のトークンを読み込んで、要素の種類を判断
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to read first token: %w", err)
	}

	startElement, ok := token.(xml.StartElement)
	if !ok {
		return nil, fmt.Errorf("expected start element, got %T", token)
	}

	switch startElement.Name.Local {
	case "create_file":
		var cf CreateFile
		if err := decoder.DecodeElement(&cf, &startElement); err != nil {
			return nil, fmt.Errorf("failed to decode create_file: %w", err)
		}
		cf.Path = strings.TrimPrefix(cf.Path, "/")
		return NewChangeFile(cf), nil
	case "modify_file":
		var mf ModifyFile
		if err := decoder.DecodeElement(&mf, &startElement); err != nil {
			return nil, fmt.Errorf("failed to decode modify_file: %w", err)
		}
		mf.Path = strings.TrimPrefix(mf.Path, "/")
		return NewChangeFile(mf), nil
	case "rename_file":
		var rf RenameFile
		if err := decoder.DecodeElement(&rf, &startElement); err != nil {
			return nil, fmt.Errorf("failed to decode rename_file: %w", err)
		}
		rf.OldPath = strings.TrimPrefix(rf.OldPath, "/")
		rf.NewPath = strings.TrimPrefix(rf.NewPath, "/")
		return NewChangeFile(rf), nil
	case "find_file":
		var ff FindFile
		if err := decoder.DecodeElement(&ff, &startElement); err != nil {
			return nil, fmt.Errorf("failed to decode find_file: %w", err)
		}
		ff.Path = strings.TrimPrefix(ff.Path, "/")
		return ff, nil
	case "attempt_complete":
		var ac AttemptComplete
		if err := decoder.DecodeElement(&ac, &startElement); err != nil {
			return nil, fmt.Errorf("failed to decode attempt_complete: %w", err)
		}
		return ac, nil
	case "create_proposal":
		var cp CreateProposal
		if err := decoder.DecodeElement(&cp, &startElement); err != nil {
			return nil, fmt.Errorf("failed to decode create_proposal: %w", err)
		}
		return cp, nil
	case "update_proposal":
		var up UpdateProposal
		if err := decoder.DecodeElement(&up, &startElement); err != nil {
			return nil, fmt.Errorf("failed to decode update_proposal: %w", err)
		}
		return up, nil
	case "query_rag":
		var qr QueryRAG
		if err := decoder.DecodeElement(&qr, &startElement); err != nil {
			return nil, fmt.Errorf("failed to decode query_rag: %w", err)
		}
		return qr, nil
	default:
		return nil, fmt.Errorf("unknown command type: %s", startElement.Name.Local)
	}
}
