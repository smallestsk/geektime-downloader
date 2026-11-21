package comments

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/nicoxiang/geektime-downloader/internal/geektime"
)

// CommExtension ...
const CommExtension = ".comments.md"

// convert to markdown string
func ConvertToMarkdown(commentList []geektime.ArticleComments) string {
	var b strings.Builder
	b.WriteString("\n\n---\n# 评论区\n\n")
	for _, resp := range commentList {
		for _, comment := range resp.CommentList {
			ctime := time.Unix(int64(comment.CommentCtime), 0).Format("2006-01-02 15:04")
			b.WriteString(fmt.Sprintf("**%s** (%s)\n", comment.UserName, ctime))
			b.WriteString(html.UnescapeString(comment.CommentContent))
			b.WriteString("\n\n")
			if len(comment.Replies) > 0 {
				for _, reply := range comment.Replies {
					replyCtime := time.Unix(int64(reply.CTime), 0).Format("2006-01-02 15:04")
					b.WriteString(fmt.Sprintf("> **%s** (%s)\n", reply.UserName, replyCtime))
					b.WriteString("> " + strings.ReplaceAll(html.UnescapeString(reply.Content), "\n", "\n> "))
					b.WriteString("\n>\n")
				}
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}

