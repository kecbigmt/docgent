package lib

import (
	"encoding/json"
	"fmt"
)

type File struct {
	Name               string              `json:"name"`
	DisplayName        string              `json:"displayName"`
	Description        string              `json:"description,omitempty"`
	CreateTime         string              `json:"createTime,omitempty"`
	UpdateTime         string              `json:"updateTime,omitempty"`
	FileStatus         *FileStatus         `json:"fileStatus"`
	DirectUploadSource *DirectUploadSource `json:"directUploadSource,omitempty"`
	GcsSource          *GcsSource          `json:"gcsSource,omitempty"`
	GoogleDriveSource  *GoogleDriveSource  `json:"googleDriveSource,omitempty"`
}

type DirectUploadSource struct{}

type GcsSource struct {
	Uris []string `json:"uris"`
}

/**
 * Google Drive
 */

type GoogleDriveSource struct {
	ResourceIds []GoogleDriveResourceID `json:"resourceIds"`
}

type GoogleDriveResourceID struct {
	ResourceType GoogleDriveResourceType `json:"resourceType"`
	ResourceID   string                  `json:"resourceId"`
}

type GoogleDriveResourceType int

const (
	GoogleDriveResourceTypeUnspecified GoogleDriveResourceType = iota
	GoogleDriveResourceTypeFile
	GoogleDriveResourceTypeFolder
)

func (s GoogleDriveResourceType) String() string {
	return []string{"RESOURCE_TYPE_UNSPECIFIED", "RESOURCE_TYPE_FILE", "RESOURCE_TYPE_FOLDER"}[s]
}

func (s GoogleDriveResourceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *GoogleDriveResourceType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch str {
	case "RESOURCE_TYPE_UNSPECIFIED":
		*s = GoogleDriveResourceTypeUnspecified
	case "RESOURCE_TYPE_FILE":
		*s = GoogleDriveResourceTypeFile
	case "RESOURCE_TYPE_FOLDER":
		*s = GoogleDriveResourceTypeFolder
	default:
		return fmt.Errorf("invalid google drive resource type: %s", str)
	}
	return nil
}

/**
 * File Status
 */

type FileStatus struct {
	State       FileState `json:"state,omitempty"`
	ErrorStatus string    `json:"errorStatus,omitempty"`
}

type FileState int

const (
	FileStateStateUnspecified FileState = iota
	FileStateActive
	FileStateError
)

func (s FileState) String() string {
	return []string{"STATE_UNSPECIFIED", "ACTIVE", "ERROR"}[s]
}

func (s FileState) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *FileState) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch str {
	case "STATE_UNSPECIFIED":
		*s = FileStateStateUnspecified
	case "ACTIVE":
		*s = FileStateActive
	case "ERROR":
		*s = FileStateError
	default:
		return fmt.Errorf("invalid file state: %s", str)
	}
	return nil
}
