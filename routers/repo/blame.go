// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	//"code.gitea.io/git"
	gotemplate "html/template"

	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/git"
	"fmt"
	"bytes"
	"code.gitea.io/gitea/modules/highlight"
	"strings"
)

const (
	tplBlame base.TplName = "repo/blame"
)

// Blame shows git-blame
func Blame(ctx *context.Context) {
	fileName := ctx.Repo.TreePath

	_, err := ctx.Repo.Commit.GetTreeEntryByPath(ctx.Repo.TreePath)
	if err != nil {
		ctx.NotFoundOrServerError("Repo.Commit.GetTreeEntryByPath", git.IsErrNotExist, err)
		return
	}

	if len(fileName) == 0 {
		ctx.Handle(404, "TreePath not foung", nil)
		return
	}
	branchName := ctx.Repo.BranchName
	blame, err := ctx.Repo.Repository.GitBlame(branchName, fileName)

	ctx.Data["RequireHighlightJS"] = true
	ctx.Data["HighlightClass"] = highlight.FileNameToHighlightClass(fileName)
	if err != nil {
		ctx.Handle(500, "FileBlame", err)
		return
	}
	var output bytes.Buffer
	for index, line := range blame.Lines {
		output.WriteString(fmt.Sprintf(`<li class="L%d" rel="L%d">%s</li>`, index+1, index+1, gotemplate.HTMLEscapeString(line.Content)) + "\n")
	}

	var treeNames []string
	paths := make([]string, 0, 5)
	if len(ctx.Repo.TreePath) > 0 {
		treeNames = strings.Split(ctx.Repo.TreePath, "/")
		for i := range treeNames {
			paths = append(paths, strings.Join(treeNames[:i+1], "/"))
		}
	}

	ctx.Data["TreeNames"] = treeNames
	ctx.Data["FileContent"] = gotemplate.HTML(output.String())
	ctx.Data["Blame"] = blame
	ctx.Data["FileName"] = fileName
	ctx.HTML(200, tplBlame)
}
