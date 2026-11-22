package geektime

import (
	"strconv"
	"time"
	"sort"

	"github.com/go-resty/resty/v2"
	"github.com/nicoxiang/geektime-downloader/internal/geektime/response"
)

const (
	// DefaultBaseURL ...
	DefaultBaseURL = "https://time.geekbang.org"

	// V1ColumnArticlesPath get all articles summary info in one column
	V1ColumnArticlesPath = "/serv/v1/column/articles"
	// V1ArticlePath used in normal column
	V1ArticlePath = "/serv/v1/article"
	// V4CommentsPath used in normal column
	V4CommentsPath = "/serv/v4/comment/list"
	// V3ColumnInfoPath used in get normal column/video info
	V3ColumnInfoPath = "/serv/v3/column/info"
	// V3ProductInfoPath used in get daily lesson, qconplus product info
	V3ProductInfoPath = "/serv/v3/product/info"
	// V3ArticleInfoPath used in normal video, daily lesson, qconplus
	V3ArticleInfoPath = "/serv/v3/article/info"
	// V3VideoPlayAuthPath used in normal video, daily lesson, qconplus video play auth
	V3VideoPlayAuthPath = "/serv/v3/source_auth/video_play_auth"

	// GeekBangCookieDomain ...
	GeekBangCookieDomain = ".geekbang.org"

	// GCID ...
	GCID = "GCID"
	// GCESS ...
	GCESS = "GCESS"
)

// Course ...
type Course struct {
	Access   bool
	ID       int
	Title    string
	Type     string
	IsVideo  bool
	Articles []Article
	TimeCreated time.Time
	TimeUpdated time.Time
	Unit string // 集数
	Author struct {
		Name string
		Intro string
		Brief string
	}
}


// Product ...
type Product struct {
	Access   bool
	ID       int
	Title    string
	Subtitle string
	IntroHTML string
	Type     string
	IsVideo  bool
	TimeCreated time.Time
	TimeUpdated time.Time
	Article struct {
		ID int
	}
	Author struct {
		Name string
		Intro string
		Info string
	}
}

// Article ...
type Article struct {
	AID          int
	SectionTitle string
	Title        string
	Summary string
}

type V3ArticleInfo struct {
	AiMindmap     string
	ID           int
	Type         int
	Title        string
	Ctime        int
	Video struct {
		ID       string
		Subtitles []struct {
			Url    string
			Format    string
			Name    string
		}
	}
	IsVideo         bool
	InPvip          int
	Content     string
	Attachments []struct{
			Url    string
			Filename    string
	}
}

type ArticleComments struct {
	Page struct {
		HasMore bool
		Count int
	}
	CommentList []struct {
		UserName string
		CommentCtime int
		Replies []struct {
			UserName string
			Content string
			CTime int
		}
		CommentContent string
	}
}

// CourseInfo get narmal geektime course info
func (c *Client) CourseInfo(productID int) (Course, error) {
	var p Course
	var err error
	p, err = c.columnInfo(productID)
	if err != nil {
		return p, err
	}

	var articles []Article
	articles, err = c.columnArticles(productID)
	if err != nil {
		return p, err
	}

	sort.Slice(articles, func(i, j int) bool {
		return articles[i].AID < articles[j].AID
	})

	p.Articles = articles

	return p, nil
}

// V1ArticleInfo ...
func (c *Client) V1ArticleInfo(articleID int) (response.V1ArticleResponse, error) {
	var res response.V1ArticleResponse
	r := c.newRequest(
		resty.MethodPost,
		DefaultBaseURL,
		V1ArticlePath,
		nil,
		map[string]interface{}{
			"id":                strconv.Itoa(articleID),
			"include_neighbors": true,
			"is_freelyread":     true,
			"reverse":           false,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return response.V1ArticleResponse{}, err
	}
	return res, nil
}

// V4CommentList ...
func (c *Client) V4CommentList(articleID int, prev int) (ArticleComments, error) {
	var res response.V4CommentListResponse
	r := c.newRequest(
		resty.MethodPost,
		DefaultBaseURL,
		V4CommentsPath,
		nil,
		map[string]interface{}{
			"aid":  articleID,
			"prev": prev,
			"sort": 0,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return ArticleComments{}, err
	}
	return ArticleComments(res.Data), nil
}

// ProductInfo used to get daily lesson or qconplus product info
func (c *Client) ProductInfo(productID int) (Product, error) {
	var res response.V3ProductInfoResponse
	r := c.newRequest(
		resty.MethodPost,
		DefaultBaseURL,
		V3ProductInfoPath,
		nil,
		map[string]interface{}{
			"id": productID,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return Product{}, err
	}

	return Product{
		Access:  res.Data.Info.Extra.Sub.AccessMask > 0,
		ID:      res.Data.Info.ID,
		Title:   res.Data.Info.Title,
		Subtitle: res.Data.Info.Subtitle,
		IntroHTML: res.Data.Info.IntroHTML,
		Type:    res.Data.Info.Type,
		IsVideo: res.Data.Info.IsVideo,
		TimeCreated: time.Unix(int64(res.Data.Info.Ctime), 0),
		TimeUpdated: time.Unix(int64(res.Data.Info.Utime), 0),
		Article: struct{ID int}{ID: res.Data.Info.Article.ID},
		Author: struct{Name string; Intro string; Info string}(res.Data.Info.Author),
	}, nil
}

// V3ArticleInfo used to get daily lesson or qconplus article info
func (c *Client) V3ArticleInfo(articleID int) (V3ArticleInfo, error) {
	var res response.V3ArticleInfoResponse
	r := c.newRequest(
		resty.MethodPost,
		DefaultBaseURL,
		V3ArticleInfoPath,
		nil,
		map[string]interface{}{
			"id": articleID,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return V3ArticleInfo{}, err
	}

	return V3ArticleInfo(res.Data.Info), nil
}

// VideoPlayAuth get play auth string
func (c *Client) VideoPlayAuth(articleID, sourceType int, videoID string) (string, error) {
	var res response.V3VideoPlayAuthResponse
	r := c.newRequest(
		resty.MethodPost,
		DefaultBaseURL,
		V3VideoPlayAuthPath,
		nil,
		map[string]interface{}{
			"aid":         articleID,
			"source_type": sourceType,
			"video_id":    videoID,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return "", err
	}
	return res.Data.PlayAuth, nil
}

// columnInfo get normal column info, like v3 product info
func (c *Client) columnInfo(productID int) (Course, error) {
	var res response.V3ColumnInfoResponse
	r := c.newRequest(
		resty.MethodPost,
		DefaultBaseURL,
		V3ColumnInfoPath,
		nil,
		map[string]interface{}{
			"product_id":             productID,
			"with_recommend_article": true,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return Course{}, err
	}

	return Course{
		Access:  res.Data.Extra.Sub.AccessMask > 0,
		ID:      res.Data.ID,
		Type:    res.Data.Type,
		Title:   res.Data.Title,
		IsVideo: res.Data.IsVideo,
		TimeCreated: time.Unix(int64(res.Data.Ctime), 0),
		TimeUpdated: time.Unix(int64(res.Data.Utime), 0),
		Unit: res.Data.Unit,
		Author: struct{Name string; Intro string; Brief string}(res.Data.Author),
	}, nil
}

// columnArticles call geektime api to get article list
func (c *Client) columnArticles(cid int) ([]Article, error) {
	res := &response.V1ColumnArticlesResponse{}
	r := c.newRequest(
		resty.MethodPost,
		DefaultBaseURL,
		V1ColumnArticlesPath,
		nil,
		map[string]interface{}{
			"cid":    strconv.Itoa(cid),
			"order":  "earliest",
			"prev":   0,
			"sample": false,
			"size":   1000, //get all articles
		},
		res,
	)
	if _, err := do(r); err != nil {
		return nil, err
	}

	var articles []Article
	for _, v := range res.Data.List {
		articles = append(articles, Article{
			AID:   v.ID,
			Title: v.ArticleTitle,
			Summary: v.ArticleSummary,
		})
	}
	return articles, nil
}

func IsTextCourse(course Course) bool {
	return !course.IsVideo
}
