package cmd

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/nicoxiang/geektime-downloader/internal/config"
	"github.com/nicoxiang/geektime-downloader/internal/fsm"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
)

var (
	geektimeClient *geektime.Client
	cfg            config.AppConfig
)

func init() {
	userHomeDir, _ := os.UserHomeDir()
	defaultDownloadFolder := filepath.Join(userHomeDir, config.GeektimeDownloaderFolder)

	rootCmd.Flags().StringVar(&cfg.Gcid, "gcid", "", "极客时间 cookie 值 gcid")
	rootCmd.Flags().StringVar(&cfg.Gcess, "gcess", "", "极客时间 cookie 值 gcess")
	rootCmd.Flags().StringVarP(&cfg.DownloadFolder, "folder", "f", defaultDownloadFolder, "专栏和视频课的下载目标位置")
	rootCmd.Flags().StringVarP(&cfg.Quality, "quality", "q", "sd", "下载视频清晰度(ld标清,sd高清,hd超清)")
	rootCmd.Flags().IntVar(&cfg.MaxCommentsPage, "max-page", 3, "评论最多下载几页, 默认最多3页")
	rootCmd.Flags().IntVar(&cfg.Interval, "interval", 1, "下载资源的间隔时间, 单位为秒, 默认1秒")
	rootCmd.Flags().BoolVar(&cfg.IsEnterprise, "enterprise", false, "是否下载企业版极客时间资源")
	rootCmd.Flags().BoolVar(&cfg.SkipEmbedVideo, "skip-embed-video", false, "是否跳过下载图文课程中的视频")
	rootCmd.Flags().BoolVar(&cfg.DownloadImages, "download-images", false, "下载图片资源到本地")
	rootCmd.Flags().IntVar(&cfg.ProductType, "product-type", 0, "产品类型，指定quiet时才生效(企业版时：0-训练营；   非企业版时：0-普通课程，1-每日一课，2-公开课，3-大厂案例，4-训练营，5-其他)")
	rootCmd.Flags().StringVar(&cfg.ProductID, "product-id", "", "产品ID/课程ID。只在指定quiet参数时有效。如需下载多个产品/课程，则以逗号分隔，例 123,234")
	rootCmd.Flags().StringVar(&cfg.ArticleID, "article-id", "", "文章ID。只在指定quiet参数时有效。若指定了文章ID，则 productID 参数只能是单值，不允许为多个值。如需下载多个文章，则以逗号分隔，例 123,234")
	rootCmd.Flags().BoolVar(&cfg.Quiet, "quiet", false, "静默模式（产品类型由product-type指定，默认下载所有文章）")
	rootCmd.Flags().StringVar(&cfg.LogLevel, "log-level", "info", "日志记录级别(debug, info, warn, error, none)")

	rootCmd.MarkFlagsRequiredTogether("gcid", "gcess")
}

var rootCmd = &cobra.Command{
	Use:          "geektime-downloader",
	Short:        "Geektime-downloader is used to download geek time lessons",
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		logger.Init(cfg.LogLevel)
		return config.ValidateConfig(&cfg)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		readCookies := config.ReadCookiesFromInput(&cfg)

		geektimeClient = geektime.NewClient(readCookies)

		runner := fsm.NewFSMRunner(cmd.Context(), &cfg, geektimeClient)
		return runner.Run()
	},
}

// Execute ...
func Execute() {
	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
