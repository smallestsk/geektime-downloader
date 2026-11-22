package comments

import (
	"fmt"
	"html"
	"os"
	"strings"
	"time"

	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
)

func FetchCommentsMD(client *geektime.Client, aid int, maxPage int) (string, error) {
	commentList, err := fetchComments(client,
		aid,
		maxPage,
	)
	if err != nil {
		logger.Errorf(err, "Failed to fetch comments, articleID: %d", aid)
		fmt.Fprintf(os.Stderr, "Failed to fetch comments, articleID: %d", aid)
		return "", err
	}

	return convertToMarkdown(commentList), nil
}

func fetchComments(client *geektime.Client, aid int, maxPage int) ([]geektime.ArticleComments, error) {
	prev := 0
	articleComments, err := client.V4CommentList(aid, prev)
	if err != nil {
		return nil, err
	}

	retList := []geektime.ArticleComments{}
	retList = append(retList, articleComments)

	// 在网站抓的报文，看起来很奇怪，没有 prev=1 的报文。所以这里也就这么处理了（跟网站抓的报文保持一致）
	prev = 2
	for articleComments.Page.HasMore && prev <= maxPage {
		articleComments, err = client.V4CommentList(aid, prev)
		if err != nil {
			return nil, err
		}
		retList = append(retList, articleComments)
		prev++
	}

	return retList, nil
}

// convert to markdown string
func convertToMarkdown(commentList []geektime.ArticleComments) string {
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

