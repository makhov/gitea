package models

import (
	"code.gitea.io/git"
	"fmt"
	"github.com/Unknwon/com"
	"time"
	"bytes"
	"regexp"
)

// Blame represents a Git object.
type BlameLine struct {
	Commit  *git.Commit
	Author  string // Signature?
	When    time.Time
	Num     string
	Content string
}

// Blame represents a Git object.
type Blame struct {
	File  string
	Lines []*BlameLine
}

// GitBlame
func (repo *Repository) GitBlame(branch, fileName string) (*Blame, error) {
	repoWorkingPool.CheckIn(com.ToStr(repo.ID))
	defer repoWorkingPool.CheckOut(com.ToStr(repo.ID))

	if err := repo.DiscardLocalRepoBranchChanges(branch); err != nil {
		return nil, fmt.Errorf("DiscardLocalRepoBranchChanges [branch: %s]: %v", branch, err)
	} else if err = repo.UpdateLocalCopyBranch(branch); err != nil {
		return nil, fmt.Errorf("UpdateLocalCopyBranch [branch: %s]: %v", branch, err)
	}

	localPath := repo.LocalCopyPath()

	gitRepo, err := git.OpenRepository(repo.RepoPath())
	if err != nil {
		return nil, fmt.Errorf("git.OpenRepository [branch: %s]: %v", branch, err)
	}

	blameBytes, err := gitRepo.FileBlame(branch, localPath, fileName)
	if err != nil {
		return nil, fmt.Errorf("git.FileBlame [branch: %s, file %s]: %v", branch, fileName, err)
	}

	return repo.parseBlameOutput(blameBytes)
}

func (repo *Repository) parseBlameOutput(blameBytes []byte) (*Blame, error) {
	b := new(Blame)
	if len(blameBytes) == 0 {
		return b, nil
	}

	gitRepo, err := git.OpenRepository(repo.RepoPath())
	if err != nil {
		return nil, fmt.Errorf("git.OpenRepository: %v", err)
	}

	parts := bytes.Split(blameBytes, []byte{'\n'})

	r := regexp.MustCompile(`^(.{8})\s\((.+)\s(\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}\s\+\d{4})\s(\d+)\)\s(.*)`)
	for _, part := range parts[0:len(parts) - 1] { // last one is empty string
		matches := r.FindAllStringSubmatch(string(part), -1)
		if len(matches) > 0 {
			commitID := matches[0][1]
			author := matches[0][2]
			dateString := matches[0][3]
			lineNum := matches[0][4]
			content := matches[0][5]

			date, err := time.Parse("2006-01-02 15:04:05 -0700", dateString)
			if err != nil {
				return nil, err
			}
			commit, err := gitRepo.GetCommit(commitID)
			if err != nil {
				return nil, err
			}
			bl := &BlameLine{
				Commit:  commit,
				Author:  author,
				When:    date,
				Num:     lineNum,
				Content: content,
			}

			b.Lines = append(b.Lines, bl)
		}
	}

	return b, nil
}
