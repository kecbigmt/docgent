package github

import (
	"docgent/internal/domain/data"
	"fmt"
	"regexp"
	"strconv"
)

type IssueCommentRef struct {
	owner           string
	repo            string
	prNumber        int
	sourceCommentID int64
}

func NewIssueCommentRef(owner, repo string, prNumber int, sourceCommentID int64) *IssueCommentRef {
	return &IssueCommentRef{owner: owner, repo: repo, prNumber: prNumber, sourceCommentID: sourceCommentID}
}

func (u *IssueCommentRef) ToURI() data.URI {
	rawURI := fmt.Sprintf("https://github.com/%s/%s/pull/%d#issuecomment-%d", u.owner, u.repo, u.prNumber, u.sourceCommentID)
	return data.NewURIUnsafe(rawURI)
}

func (u *IssueCommentRef) Owner() string {
	return u.owner
}

func (u *IssueCommentRef) Repo() string {
	return u.repo
}

func (u *IssueCommentRef) PRNumber() int {
	return u.prNumber
}

func (u *IssueCommentRef) SourceCommentID() int64 {
	return u.sourceCommentID
}

// https://github.com/{owner}/{repo}/pull/{prNumber}#issuecomment-{sourceCommentID}
var re = regexp.MustCompile(`^https://github\.com/([^/]+)/([^/]+)/pull/([^/]+)#issuecomment-([^/]+)$`)

func ParseIssueCommentRef(uri data.URI) (*IssueCommentRef, error) {
	rawURI := uri.String()
	matches := re.FindStringSubmatch(rawURI)
	if len(matches) == 5 {
		prNumber, err := strconv.Atoi(matches[3])
		if err != nil {
			return nil, fmt.Errorf("invalid PR number: %s", matches[3])
		}
		sourceCommentID, err := strconv.ParseInt(matches[4], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid source comment ID: %s", matches[4])
		}
		return NewIssueCommentRef(matches[1], matches[2], prNumber, sourceCommentID), nil
	}
	return nil, fmt.Errorf("invalid URI: %s", uri)
}
