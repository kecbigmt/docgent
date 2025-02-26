package tooluse

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// Parse parses XML string and returns Union
func Parse(xmlStr string) (Union, error) {
	// 空のXMLをパースしない
	if strings.TrimSpace(xmlStr) == "" {
		return nil, fmt.Errorf("empty xml string")
	}

	// XMLのルート要素名を取得
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get first token: %w", err)
	}

	startElement, ok := token.(xml.StartElement)
	if !ok {
		return nil, fmt.Errorf("first token is not start element")
	}

	switch startElement.Name.Local {
	case "create_file":
		var cf CreateFile
		if err := xml.Unmarshal([]byte(xmlStr), &cf); err != nil {
			return nil, fmt.Errorf("failed to unmarshal create_file: %w", err)
		}
		return NewChangeFile(cf), nil
	case "modify_file":
		var mf ModifyFile
		if err := xml.Unmarshal([]byte(xmlStr), &mf); err != nil {
			return nil, fmt.Errorf("failed to unmarshal modify_file: %w", err)
		}
		return NewChangeFile(mf), nil
	case "rename_file":
		var rf RenameFile
		if err := xml.Unmarshal([]byte(xmlStr), &rf); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rename_file: %w", err)
		}
		return NewChangeFile(rf), nil
	case "delete_file":
		var df DeleteFile
		if err := xml.Unmarshal([]byte(xmlStr), &df); err != nil {
			return nil, fmt.Errorf("failed to unmarshal delete_file: %w", err)
		}
		return NewChangeFile(df), nil
	case "find_file":
		var ff FindFile
		if err := xml.Unmarshal([]byte(xmlStr), &ff); err != nil {
			return nil, fmt.Errorf("failed to unmarshal find_file: %w", err)
		}
		return ff, nil
	case "create_proposal":
		var cp CreateProposal
		if err := xml.Unmarshal([]byte(xmlStr), &cp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal create_proposal: %w", err)
		}
		return cp, nil
	case "update_proposal":
		var up UpdateProposal
		if err := xml.Unmarshal([]byte(xmlStr), &up); err != nil {
			return nil, fmt.Errorf("failed to unmarshal update_proposal: %w", err)
		}
		return up, nil
	case "link_sources":
		var ls LinkSources
		if err := xml.Unmarshal([]byte(xmlStr), &ls); err != nil {
			return nil, fmt.Errorf("failed to unmarshal link_sources: %w", err)
		}
		return ls, nil
	case "attempt_complete":
		var ac AttemptComplete
		if err := xml.Unmarshal([]byte(xmlStr), &ac); err != nil {
			return nil, fmt.Errorf("failed to unmarshal attempt_complete: %w", err)
		}
		return ac, nil
	case "query_rag":
		var qr QueryRAG
		if err := xml.Unmarshal([]byte(xmlStr), &qr); err != nil {
			return nil, fmt.Errorf("failed to unmarshal query_rag: %w", err)
		}
		return qr, nil
	case "find_source":
		var fs FindSource
		if err := xml.Unmarshal([]byte(xmlStr), &fs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal find_source: %w", err)
		}
		return fs, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", startElement.Name.Local)
	}
}
