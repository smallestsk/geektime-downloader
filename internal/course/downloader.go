package course

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"golang.org/x/net/html"

	"github.com/nicoxiang/geektime-downloader/internal/comments"
	"github.com/nicoxiang/geektime-downloader/internal/config"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/markdown"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/filenamify"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/files"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
	"github.com/nicoxiang/geektime-downloader/internal/ui"
	"github.com/nicoxiang/geektime-downloader/internal/video"
)

const (
)

type CourseDownloader struct {
	ctx                context.Context
	cfg                *config.AppConfig
	geektimeClient     *geektime.Client
	concurrency        int
	waitRand           *rand.Rand
	downloadingSpinner *spinner.Spinner
}

func NewCourseDownloader(ctx context.Context, cfg *config.AppConfig, geektimeClient *geektime.Client, sp *spinner.Spinner) *CourseDownloader {
	concurrency := int(math.Ceil(float64(runtime.NumCPU()) / 2.0))
	if concurrency <= 0 {
		concurrency = 1
	}
	return &CourseDownloader{
		ctx:                ctx,
		cfg:                cfg,
		geektimeClient:     geektimeClient,
		concurrency:        concurrency,
		waitRand:           rand.New(rand.NewSource(time.Now().UnixNano())),
		downloadingSpinner: sp,
	}
}

func saveProductInfoFile(course geektime.Product, dir string, comments string) error {
	courseInfoFileName := path.Join(dir, filenamify.Filenamify("00-geektime课程简介") + ".md")
	// if files.CheckFileExists(courseInfoFileName) { // skip if file already exists.
	// 	return nil
	// }

	// ---
	// date created: time.Now()
	// course title: $course.title
	// author: $course.Author.Name
	// course created date: $course.TimeCreated.Format("2020/01/30 13:00:00")
	// course updated date: $course.TimeUpdated.Format("2020/01/30 13:00:00")
	// ---
	//
	frontMatter := fmt.Sprintf("---\n" +
							"date created: %s\n" +
							"course title: %s\n" +
							"author: %s\n" +
							"course created date: %s\n" +
							"course updated date: %s\n" +
							"---\n\n",
						time.Now().Format("2006-01-02 15:04"),
						course.Title,
						course.Author.Name,
						course.TimeCreated.Format("2006-01-02 15:04"),
						course.TimeUpdated.Format("2006-01-02 15:04"),
					)

	// # Course info
	// Title: $course.Title
	// Title: $course.Subtitle
	// course created date: $course.TimeCreated.Format("2020/01/30 13:00:00")
	// course updated date: $course.TimeUpdated.Format("2020/01/30 13:00:00")
	//
	courseInfo := fmt.Sprintf("\n\n# Course info\n" +
							"course id: %d\n" +
							"Title: %s\n" +
							"Subtitle: %s\n" +
							"course created date: %s\n" +
							"course updated date: %s\n\n",
						course.ID,
						course.Title,
						course.Subtitle,
						course.TimeCreated.Format("2006-01-02 15:04"),
						course.TimeUpdated.Format("2006-01-02 15:04"),
					)

	// # Author info
	// author name: $course.Author.Name
	// author introduction: $course.Author.Intro
	// author brief:
	// ```
	// $course.Author.Brief
	// ```
	//
	authorInfo := fmt.Sprintf("# Author info\n" +
							"author name: %s\n" +
							"author introduction: %s\n" +
							"author info: \n" +
							"```\n" +
							"%s\n" +
							"```\n",
						course.Author.Name,
						html.UnescapeString(course.Author.Intro),
						html.UnescapeString(course.Author.Info),
					)
	
	productIntro, err := markdown.ConvertHtmlToMarkdown(course.IntroHTML)
	if err != nil {
		return err
	}

	// // 保存 course info 到文件
	f, err := os.Create(courseInfoFileName)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return err
	}
	// step3: write comments file
	_, err = f.WriteString(frontMatter + productIntro + courseInfo + authorInfo + comments)
	if err != nil {
		return err
	}
	logger.Infof("Finish download product info file, aid:%s, title: %s", course.ID, course.Title)
	return nil
}

func saveCourseInfoFile(course geektime.Course, dir string) error {
	courseInfoFileName := path.Join(dir, filenamify.Filenamify("00-geektime课程简介") + ".md")
	// if files.CheckFileExists(courseInfoFileName) { // skip if file already exists.
	// 	return nil
	// }

	// ---
	// date created: time.Now()
	// course created date: $course.TimeCreated.Format("2020/01/30 13:00:00")
	// course updated date: $course.TimeUpdated.Format("2020/01/30 13:00:00")
	// author: $course.Author.Name
	// ---
	//
	frontMatter := fmt.Sprintf("---\n" +
							"date created: %s\n" +
							"course created date: %s\n" +
							"course updated date: %s\n" +
							"author: %s\n" +
							"\n---\n\n",
						time.Now().Format("2006-01-02 15:04"),
						course.TimeCreated.Format("2006-01-02 15:04"),
						course.TimeUpdated.Format("2006-01-02 15:04"),
						course.Author.Name,
					)

	// # Course info
	// Title: $course.Title
	// course created date: $course.TimeCreated.Format("2020/01/30 13:00:00")
	// course updated date: $course.TimeUpdated.Format("2020/01/30 13:00:00")
	// course unit count: $course.Unit
	//
	courseInfo := fmt.Sprintf("# Course info\n" +
							"course id: %d\n" +
							"Title: %s\n" +
							"course created date: %s\n" +
							"course updated date: %s\n" +
							"course unit count: %s\n\n",
						course.ID,
						course.Title,
						course.TimeCreated.Format("2006-01-02 15:04"),
						course.TimeUpdated.Format("2006-01-02 15:04"),
						course.Unit,
					)

	// # Author info
	// author name: $course.Author.Name
	// author introduction: $course.Author.Intro
	// author brief:
	// ```
	// $course.Author.Brief
	// ```
	//
	authorInfo := fmt.Sprintf("# Author info\n" +
							"author name: %s\n" +
							"author introduction: %s\n" +
							"author brief: \n" +
							"```\n" +
							"%s\n" +
							"```\n\n",
						course.Author.Name,
						html.UnescapeString(course.Author.Intro),
						html.UnescapeString(course.Author.Brief),
					)

	// # Article list
	// | chapter | article |
	// | -- | -- |
	// | [[$listItem.Title]] |
	var b strings.Builder
	for _, article := range course.Articles {
		b.WriteString(fmt.Sprintf("| %s | [[%s]] |\n", filenamify.Filenamify(article.ChapterTitle), filenamify.Filenamify(article.Title)))
	}
	articleTable := "# Article list\n\n" +
					"| chapter | article |\n" +
					"| --- | --- |\n" +
					b.String()

	// // 保存 course info 到文件
	f, err := os.Create(courseInfoFileName)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return err
	}
	// step3: write comments file
	_, err = f.WriteString(frontMatter + courseInfo + authorInfo + articleTable)
	if err != nil {
		return err
	}
	logger.Infof("Finish download course info file, aid:%s, title: %s", course.ID, course.Title)
	return nil
}

// DownloadAll manages the bulk download process for all articles in a selected product (course).
// Returns an error if any step in the download process fails.
func (d *CourseDownloader) DownloadAll(course geektime.Course, productType ui.ProductTypeSelectOption) error {
	columnDir, err := d.mkDownloadColumnDir(course.Title)
	if err != nil {
		return err
	}

	if err := saveCourseInfoFile(course, columnDir); err != nil {
		fmt.Fprintf(os.Stderr,"课程信息下载失败， aid: %d, title: %s\n", course.ID, course.Title)
		return err
	}

	fmt.Printf("正在下载专栏 %d-《%s》 中的所有文章\n", course.ID, course.Title)
	if geektime.IsTextCourse(course) {
		total := len(course.Articles)
		var downloaded int

		for _, article := range course.Articles {
			err := d.DownloadArticle(course, productType, article, false, true)
			if err != nil {
				fmt.Fprintf(os.Stderr, "文章下载失败， aid: %d, title: %s\n", article.AID, article.Title)
				return err
			}
			increaseDownloadedTextArticleCount(total, &downloaded)
		}
	} else {
		for _, article := range course.Articles {
			err := d.DownloadArticle(course, productType, article, false, true)
			if err != nil {
				fmt.Fprintf(os.Stderr, "文章下载失败， aid: %d, title: %s\n", article.AID, article.Title)
				return err
			}
		}
	}

	fmt.Printf("\n")
	return nil
}

// DownloadArticle processes the download of a single article from Geektime.
// It handles both text-based courses and video content differently.
func (d *CourseDownloader) DownloadArticle(course geektime.Course, productType ui.ProductTypeSelectOption, article geektime.Article, overwrite bool, waitRandomTime bool) error {
	columnDir, err := d.mkDownloadColumnDir(course.Title)
	if err != nil {
		return err
	}

	if geektime.IsTextCourse(course) {
		skip := skipDownloadTextArticle(article, columnDir, overwrite)
		if skip {
			return nil
		}
		if waitRandomTime {
			d.waitRandomTime()
		}

		d.downloadingSpinner.Prefix = fmt.Sprintf("[ 正在下载 《%s》, articleID: %d... ]", article.Title, article.AID)
		d.downloadingSpinner.Start()
		defer d.downloadingSpinner.Stop()
		return d.downloadTextArticle(article, columnDir, overwrite)
	} else {
		skip := skipDownloadVideoArticle(article, columnDir, overwrite)

		// 训练营特殊处理，训练营只能从文章详情中获取当前文章是否是视频，训练营目前只支持下载视频类文章，
		// 下载所有时如果是文本类，直接跳过
		if productType.IsUniversity() {
			universityArticleDetail, err := d.geektimeClient.UniversityClassArticleDetail(course.ID, article.AID)
			if err != nil {
				return err
			}
			if universityArticleDetail.Data.VideoID == "" {
				skip = true
			}
		}
		if skip {
			return nil
		}
		if waitRandomTime {
			d.waitRandomTime()
		}

		return d.downloadVideoArticle(course, productType, article, columnDir)
	}
}

// DownloadSingleVideoProduct downloads a single video product.
// 每日一课，大厂案例等
// 大厂案例课是 Product 类型，然后 Product 下挂 Article， Article 内有一个 video。
// comments 也是跟 Article 绑定的（用 articleID 获取评论列表）
func (d *CourseDownloader) DownloadSingleVideoProduct(productInfo geektime.Product, sourceType int) error {
	columnDir, err := d.mkDownloadColumnDir(productInfo.Title)
	if err != nil {
		return err
	}

	commentsMDString, _ := comments.FetchCommentsMD(d.geektimeClient, productInfo.Article.ID, d.cfg.MaxCommentsPage)

	if err := saveProductInfoFile(productInfo, columnDir, commentsMDString); err != nil {
		fmt.Printf("课程信息下载失败， productID: %d, title: %s\n", productInfo.ID, productInfo.Title)
		return err
	}

	articleInfo, err := d.geektimeClient.V3ArticleInfo(productInfo.Article.ID)
	if err != nil {
		return err
	}

	return video.DownloadArticleVideo(d.ctx, d.geektimeClient, articleInfo, sourceType, columnDir, d.cfg.Quality, d.concurrency)
}

func increaseDownloadedTextArticleCount(total int, downloaded *int) {
	current := *downloaded + 1
	*downloaded = current
	if current > total {
		current = total
	}
	fmt.Printf("\r已完成下载%d/%d", current, total)
}

func shouldSkipFile(filepath string, overwrite bool) bool {
	if overwrite {
		return false
	}

	if !files.CheckFileExists(filepath) {
		return false
	}

	return true
}

func skipDownloadTextArticle(article geektime.Article, columnDir string, overwrite bool) bool {
	markdownFilePath := filepath.Join(columnDir, filenamify.Filenamify(article.Title) + markdown.MDExtension)
	return shouldSkipFile(markdownFilePath, overwrite)
}

// downloadTextArticle downloads the content of a Geektime text article in various formats (PDF, Markdown, Audio, and Video).
// The function supports overwriting existing files if specified.
func (d *CourseDownloader) downloadTextArticle(article geektime.Article, columnDir string, overwrite bool) error {
	articleInfo, err := d.geektimeClient.V1ArticleInfo(article.AID)
	if err != nil {
		return err
	}

	if !d.cfg.SkipEmbedVideo {
		// video exist in article content
		hasVideo, videoURL := getVideoURLFromArticleContent(articleInfo.Data.ArticleContent)
		if hasVideo && videoURL != "" {
			if err := video.DownloadMP4(d.ctx, article.Title, columnDir, []string{videoURL}, overwrite); err != nil {
				return err
			}
		}

		// inline video
		if len(articleInfo.Data.InlineVideoSubtitles) > 0 {
			videoURLs := make([]string, len(articleInfo.Data.InlineVideoSubtitles))
			for i, v := range articleInfo.Data.InlineVideoSubtitles {
				videoURLs[i] = v.VideoURL
			}
			if err := video.DownloadMP4(d.ctx, article.Title, columnDir, videoURLs, overwrite); err != nil {
				return err
			}
		}
	}

	commentsMDString, _ := comments.FetchCommentsMD(d.geektimeClient, article.AID, d.cfg.MaxCommentsPage)

	if err := markdown.Download(d.ctx,
		time.Unix(int64(articleInfo.Data.ArticleCtime), 0),
		articleInfo.Data.ArticleContent,
		article.Title,
		commentsMDString,
		columnDir,
		article.AID,
		articleInfo.Data.AiMindmap,
		d.cfg.DownloadImages,
	); err != nil {
		return err
	}

	return nil
}

func skipDownloadVideoArticle(article geektime.Article, columnDir string, overwrite bool) bool {
	videoFilePath := filepath.Join(columnDir, filenamify.Filenamify(article.Title) + video.TSExtension)
	return shouldSkipFile(videoFilePath, overwrite)
}

// downloadVideoArticle downloads a video article to the specified column directory.
// It handles different types of video content including university courses, enterprise content,
// and regular article videos.
func (d *CourseDownloader) downloadVideoArticle(course geektime.Course, productType ui.ProductTypeSelectOption, article geektime.Article, columnDir string) error {
	dir := columnDir
	var err error
	if article.SectionTitle != "" {
		dir, err = d.mkDownloadProjectSectionDir(columnDir, article.SectionTitle)
		if err != nil {
			return err
		}
	}

	if productType.IsUniversity() {
		err = video.DownloadUniversityVideo(d.ctx, d.geektimeClient, article.AID, course, dir, d.cfg.Quality, d.concurrency)
	} else if d.cfg.IsEnterprise {
		err = video.DownloadEnterpriseArticleVideo(d.ctx, d.geektimeClient, article.AID, dir, d.cfg.Quality, d.concurrency)
	} else {
		
		articleInfo, err := d.geektimeClient.V3ArticleInfo(article.AID)
		if err != nil {
			return err
		}

		// 下载文章信息，包含文章内容，评论
		commentsMDString, _ := comments.FetchCommentsMD(d.geektimeClient, article.AID, d.cfg.MaxCommentsPage)

		if err := markdown.Download(d.ctx,
					time.Unix(int64(articleInfo.Ctime), 0),
					articleInfo.Content,
					articleInfo.Title,
					commentsMDString,
					dir,
					articleInfo.ID,
					articleInfo.AiMindmap,
					false,
		); err != nil {
			fmt.Fprintf(os.Stderr, "文章信息下载失败， articleID: %d, article Title: %s\n", articleInfo.ID, articleInfo.Title)
			logger.Errorf(err, "文章信息下载失败， articleID: %d, article Title: %s", articleInfo.ID, articleInfo.Title)
		}

		err = video.DownloadArticleVideo(d.ctx, d.geektimeClient, articleInfo, productType.SourceType, dir, d.cfg.Quality, d.concurrency)
	}
	return err
}

// mkDownloadColumnDir creates a directory for downloading a column with the given columnName.
func (d *CourseDownloader) mkDownloadColumnDir(columnName string) (string, error) {
	path := filepath.Join(d.cfg.DownloadFolder, filenamify.Filenamify(columnName))
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}
	return path, nil
}

// mkDownloadProjectSectionDir creates a sub directory for project sections.
func (d *CourseDownloader) mkDownloadProjectSectionDir(projectDir, sectionName string) (string, error) {
	path := filepath.Join(projectDir, filenamify.Filenamify(sectionName))
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}
	return path, nil
}

// Sometime video exist in article content, see issue #104
// <p>
// <video poster="https://static001.geekbang.org/resource/image/6a/f7/6ada085b44eddf37506b25ad188541f7.jpg" preload="none" controls="">
// <source src="https://media001.geekbang.org/customerTrans/fe4a99b62946f2c31c2095c167b26f9c/30d99c0d-16d14089303-0000-0000-01d-dbacd.mp4" type="video/mp4">
// <source src="https://media001.geekbang.org/2ce11b32e3e740ff9580185d8c972303/a01ad13390fe4afe8856df5fb5d284a2-f2f547049c69fa0d4502ab36d42ea2fa-sd.m3u8" type="application/x-mpegURL">
// <source src="https://media001.geekbang.org/2ce11b32e3e740ff9580185d8c972303/a01ad13390fe4afe8856df5fb5d284a2-2528b0077e78173fd8892de4d7b8c96d-hd.m3u8" type="application/x-mpegURL"></video>
// </p>
func getVideoURLFromArticleContent(content string) (hasVideo bool, videoURL string) {
	if !strings.Contains(content, "<video") || !strings.Contains(content, "<source") {
		return false, ""
	}
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return false, ""
	}
	hasVideo, videoURL = false, ""
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "video" {
			hasVideo = true
		}
		if n.Type == html.ElementNode && n.Data == "source" {
			for _, a := range n.Attr {
				if a.Key == "src" && hasVideo && strings.HasSuffix(a.Val, ".mp4") {
					videoURL = a.Val
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return hasVideo, videoURL
}

// waitRandomTime wait interval seconds of time plus a 2000ms max jitter
func (d *CourseDownloader) waitRandomTime() {
	randomMillis := d.cfg.Interval*1000 + d.waitRand.Intn(2000)
	time.Sleep(time.Duration(randomMillis) * time.Millisecond)
}
