package fsm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
	"strings"
	"strconv"

	"github.com/briandowns/spinner"
	"github.com/manifoldco/promptui"

	"github.com/nicoxiang/geektime-downloader/internal/config"
	"github.com/nicoxiang/geektime-downloader/internal/course"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
	"github.com/nicoxiang/geektime-downloader/internal/ui"
)

type FSMRunner struct {
	ctx                 context.Context
	currentState        State
	config              *config.AppConfig
	selectedProductType ui.ProductTypeSelectOption
	selectedProduct     geektime.Course
	sp                  *spinner.Spinner
	geektimeClient      *geektime.Client
	courseDownloader    *course.CourseDownloader
}

// NewFSMRunner creates and initializes a new FSMRunner instance
func NewFSMRunner(ctx context.Context, cfg *config.AppConfig, geektimeClient *geektime.Client) *FSMRunner {
	sp := spinner.New(spinner.CharSets[4], 100*time.Millisecond)
	return &FSMRunner{
		ctx:              ctx,
		currentState:     StateSelectProductType,
		config:           cfg,
		sp:               sp,
		geektimeClient:   geektimeClient,
		courseDownloader: course.NewCourseDownloader(ctx, cfg, geektimeClient, sp),
	}
}

func (r *FSMRunner) Run() error {
	if r.config.Quiet {
		return r.runQuietMode()
	}
	return r.runInteractiveMode()
}

// 使用 productType 和 productIDs 下载
// 如果指定了 articleIDs，则下载指定的文章，否则下载所有文章
func (r *FSMRunner) runQuietMode() error {
	productIDs := strings.Split(r.config.ProductID, ",")
	var articleIDs []string = []string{}
	if r.config.ArticleID != "" {
		articleIDs = strings.Split(r.config.ArticleID, ",")
	}

	productType := r.config.ProductType
	var failedProductIDs []string = []string{}

	for _, productID := range productIDs {
		err := r.download(productType, productID, articleIDs)
		if err != nil {
			logger.Errorf(err, "download failed, productType: %d, productID: %s", productType, productID)
			fmt.Fprintf(os.Stderr, "download failed, productType: %d, productID: %s\n", productType, productID)
			failedProductIDs = append(failedProductIDs, productID)
		}
	}

	if len(failedProductIDs) > 0 {
		fmt.Fprintf(os.Stderr, "下载失败的 productID: %s\n", strings.Join(failedProductIDs, ","))
	}
	return nil
}

func (r *FSMRunner) download(productType int, productID string, articleIDs []string) error {
	productTypeOption := ui.GetProductTypeSelectOptionByIndex(productType, r.config.IsEnterprise)
	productIDInt, err := strconv.Atoi(productID)
	if err != nil {
		return fmt.Errorf("invalid product ID: %s", productID)
	}

	if !productTypeOption.NeedSelectArticle {
		return  r.handleInputProductIDIfDownloadDirectly(productIDInt, productTypeOption)
	} else {
		// queryCourseInfo by productID
		//     articleIDs 为空，则 r.courseDownloader.DownloadAll(courseInfo, productTypeOption)
		//     如果不为空，则 r.courseDownloader.DownloadArticle(r.selectedProduct, r.selectedProductType, a, true, false); // a 从 courseInfo.Articles 中取

		course, err := r.queryCourseInfo(productIDInt, productTypeOption)
		if err != nil {
			return err
		}

		if len(articleIDs) == 0 {
			return r.courseDownloader.DownloadAll(course, productTypeOption)
		} else {
			articleMap := make(map[int]geektime.Article, len(articleIDs))
			for _, article := range course.Articles {
				articleMap[article.AID] = article
			}

			for _, articleID := range articleIDs {
				articleIDInt, err := strconv.Atoi(articleID)
				if err != nil {
					return fmt.Errorf("invalid article ID: %s", articleID)
				}
				article, exists := articleMap[articleIDInt]
				if !exists {
					logger.Errorf(nil, "productID: %d, articleID: %d not found in course.Articles", productIDInt, articleIDInt)
					fmt.Fprintf(os.Stdout, "productID: %d, articleID: %d not found in course.Articles\n", productIDInt, articleIDInt)
					continue
				}
				err = r.downloadArticle(productTypeOption, course, article)
				if err != nil {
					return fmt.Errorf("downloadArticle failed, productID: %d, articleID: %d", productIDInt, articleIDInt)
				}
			}
		}
	}

	return nil
}

// Run executes the finite state machine loop, handling user input and state transitions.
func (r *FSMRunner) runInteractiveMode() error {
	for {
		var err error
		switch r.currentState {
		case StateSelectProductType:
			var selectedOption ui.ProductTypeSelectOption
			selectedOption, err = ui.ProductTypeSelect(r.config.IsEnterprise)
			if err == nil {
				r.currentState = StateInputProductID
				r.selectedProductType = selectedOption
			}
		case StateInputProductID:
			var productID int
			productID, err = ui.ProductIDInput(r.selectedProductType)
			if err == nil {
				if r.selectedProductType.NeedSelectArticle {
					err = r.handleInputProductIDIfNeedSelectArticle(productID)
				} else {
					err = r.handleInputProductIDIfDownloadDirectly(productID, r.selectedProductType)
				}
			}
		case StateProductAction:
			var index int
			index, err = ui.ProductAction(r.selectedProduct)
			if err == nil {
				switch index {
				case 0:
					r.currentState = StateSelectProductType
				case 1:
					err = r.handleDownloadAll()
				case 2:
					r.currentState = StateSelectArticle
				}
			}
		case StateSelectArticle:
			var index int
			index, err = ui.ArticleSelect(r.selectedProduct.Articles)
			if err == nil {
				err = r.handleSelectArticle(index, r.selectedProductType, r.selectedProduct)
			}
		}

		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				// clear line
				fmt.Print("\033[1A\033[2K")
				return nil
			case errors.Is(err, promptui.ErrInterrupt):
				// clear two lines beacause promptui print one more line if interrupt
				fmt.Print("\033[1A\033[2K\033[1A\033[2K")
				return nil
			case os.IsTimeout(err):
				logger.Errorf(err, "Request timed out")
				return fmt.Errorf("请求超时")
			default:
				logger.Errorf(err, "An error occurred")
				return err
			}
		}
	}
}

func (r *FSMRunner) queryCourseInfo(productID int, productTypeOption ui.ProductTypeSelectOption) (geektime.Course, error) {
	// choose download all or download specified article
	r.sp.Prefix = "[ 正在加载课程信息... ]"
	r.sp.Start()
	defer r.sp.Stop()

	var course geektime.Course
	var err error
	if r.selectedProductType.IsEnterpriseMode {
		// TODO: check enterprise course type
		course, err = r.geektimeClient.EnterpriseCourseInfo(productID)
		if err != nil {
			return geektime.Course{}, err
		}

	} else {
		if r.selectedProductType.IsUniversity() {
			// university don't need check product type
			// if input invalid id, access mark is 0
			course, err = r.geektimeClient.UniversityClassInfo(productID)
			if err != nil {
				return geektime.Course{}, err
			}
		} else {
			course, err = r.geektimeClient.CourseInfo(productID)
			if err == nil {
				valid := r.validateProductCode(course.Type, productTypeOption)
				// if check product type fail, re-input product
				if !valid {
					err := fmt.Errorf("(queryCourseInfo) not valid course, productID: %d, courseType: %s", productID, course.Type)
					return geektime.Course{}, err
				}
			} else {
				return geektime.Course{}, err
			}
		}
	}
	if !course.Access {
		fmt.Fprint(os.Stderr, "尚未购买该课程\n")
		return geektime.Course{}, fmt.Errorf("(queryCourseInfo) course not accessible, productID: %d, courseType: %s", productID, course.Type)
	}

	return course, nil
}

// queryCourseInfo
func (r *FSMRunner) handleInputProductIDIfNeedSelectArticle(productID int) error {
	course, err := r.queryCourseInfo(productID, r.selectedProductType)
	if err != nil {
		logger.Errorf(err, "handleInputProductIDIfNeedSelectArticle error, productID: %d", productID)
		r.currentState = StateInputProductID
		return nil
	}
	r.selectedProduct = course
	r.currentState = StateProductAction
	return nil
}

func (r *FSMRunner) handleInputProductIDIfDownloadDirectly(productID int, productTypeOption ui.ProductTypeSelectOption) error {
	// when product type is daily lesson or qconplus,
	// input id means product id
	// download video directly
	productInfo, err := r.geektimeClient.ProductInfo(productID)
	if err != nil {
		return err
	}

	if productInfo.Access == false {
		fmt.Fprint(os.Stderr, "尚未购买该课程, productID: %d\n", productID)
		r.currentState = StateInputProductID
		return nil
	}

	if r.validateProductCode(productInfo.Type, productTypeOption) {
		err = r.courseDownloader.DownloadSingleVideoProduct(productInfo,
			productTypeOption.SourceType)
		if err != nil {
			return err
		}
	}
	r.currentState = StateInputProductID
	return nil
}

// validateProductCode checks if the product code field in the response body returned by the API
// exists in the selected product's accepted product types list.
func (r *FSMRunner) validateProductCode(productCode string, productTypeOption ui.ProductTypeSelectOption) bool {
	for _, pt := range productTypeOption.AcceptProductTypes {
		if pt == productCode {
			return true
		}
	}
	fmt.Fprint(os.Stderr, "\r输入的课程 ID 有误\n")
	return false
}

func (r *FSMRunner) downloadArticle(productTypeOption ui.ProductTypeSelectOption, course geektime.Course, article geektime.Article) error {
	// 训练营特殊处理，训练营只能从文章详情中获取当前文章是否是视频，训练营目前只支持下载视频类文章，
	// 选择文章下载时，如果是文本类，提示用户重新选择
	if productTypeOption.IsUniversity() {
		universityArticleDetail, err := r.geektimeClient.UniversityClassArticleDetail(course.ID, article.AID)
		if err != nil {
			return err
		}
		if universityArticleDetail.Data.VideoID == "" {
			fmt.Printf("\r训练营暂时只支持下载视频")
			time.Sleep(time.Second)
			return fmt.Errorf("训练营暂时只支持下载视频，该文章是文本类，articleID: %d", article.AID)
		}
	}

	err := r.courseDownloader.DownloadArticle(course, productTypeOption, article, true, false)
	if err != nil {
		return err
	}
	fmt.Printf("\r%s 下载完成", article.Title)
	time.Sleep(time.Second)

	return nil
}

func (r *FSMRunner) handleSelectArticle(index int, selectedProductType ui.ProductTypeSelectOption, selectedProduct geektime.Course) error {
	if index == 0 {
		r.currentState = StateProductAction
		return nil
	}
	a := r.selectedProduct.Articles[index-1]

	// // 训练营特殊处理，训练营只能从文章详情中获取当前文章是否是视频，训练营目前只支持下载视频类文章，
	// // 选择文章下载时，如果是文本类，提示用户重新选择
	// if selectedProductType.IsUniversity() {
	// 	universityArticleDetail, err := r.geektimeClient.UniversityClassArticleDetail(selectedProduct.ID, a.AID)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if universityArticleDetail.Data.VideoID == "" {
	// 		fmt.Printf("\r训练营暂时只支持下载视频，请重新选择")
	// 		time.Sleep(time.Second)
	// 		r.currentState = StateSelectArticle
	// 		return nil
	// 	}
	// }

	// err := r.courseDownloader.DownloadArticle(r.selectedProduct, r.selectedProductType, a, true, false)
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("\r%s 下载完成", a.Title)
	// time.Sleep(time.Second)
	err := r.downloadArticle(selectedProductType, selectedProduct, a)
	if err != nil {
		logger.Errorf(err, "handleSelectArticle error, productID: %d, articleID: %d", selectedProduct.ID, a.AID)
	}
	r.currentState = StateSelectArticle
	return nil
}

// handleDownloadAll manages the bulk download process for all articles in a selected product (course).
// Returns an error if any step in the download process fails.
func (r *FSMRunner) handleDownloadAll() error {
	if err := r.courseDownloader.DownloadAll(r.selectedProduct, r.selectedProductType); err != nil {
		return err
	}
	r.currentState = StateSelectProductType
	return nil
}
