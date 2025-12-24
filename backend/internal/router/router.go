// Package router 路由配置
package router

import (
	"time"

	"feiniu-user-system/internal/config"
	"feiniu-user-system/internal/database"
	"feiniu-user-system/internal/handler"
	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/auth"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Setup 设置路由
func Setup(cfg *config.Config, logger *zap.Logger) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// 全局中间件
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.CORS())

	// 初始化数据库
	db := database.GetDB()

	// 域名白名单验证（动态从数据库读取配置）
	r.Use(middleware.DomainWhitelist(db))

	// 初始化JWT管理器
	jwtManager := auth.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpire,
		cfg.JWT.RefreshExpire,
		cfg.JWT.Issuer,
	)

	// 初始化服务
	userService := service.NewUserService(db, jwtManager, cfg)
	adminService := service.NewAdminService(db, cfg)
	memberService := service.NewMemberService(db, cfg)
	watchService := service.NewWatchService(db)
	notificationService := service.NewNotificationService(db)
	cardService := service.NewCardService(db)
	cardExportService := service.NewCardExportService(db)
	memberPaymentService := service.NewMemberPaymentService(db)
	vipPurchaseService := service.NewVipPurchaseService(db)
	wallpaperService := service.NewWallpaperService(logger)
	mediaService := service.NewMediaService(db, cfg)
	announcementService := service.NewAnnouncementService(db)
	ipBlacklistService := service.NewIPBlacklistService(db)
	healthService := service.NewHealthService(db)
	deviceService := service.NewDeviceService(db)
	importService := service.NewImportService(db)
	backupService := service.NewBackupService(cfg)
	inviteService := service.NewInviteService(db)
	statsService := service.NewStatsService(db)
	pointsService := service.NewPointsService(db)
	pointsCardService := service.NewPointsCardService(db, pointsService)
	forumService := service.NewForumService(db)
	messageService := service.NewMessageService(db)
	externalCardService := service.NewExternalCardService(db)
	goofishService := service.NewGoofishService(db, cfg.JWT.Secret) // 使用JWT密钥作为加密密钥
	alipayService := service.NewAlipayService(db, cfg.JWT.Secret)   // 支付宝服务

	// 初始化处理器
	userHandler := handler.NewUserHandler(userService)
	deviceHandler := handler.NewDeviceHandler(deviceService)
	importHandler := handler.NewImportHandler(importService)
	backupHandler := handler.NewBackupHandler(backupService)
	inviteHandler := handler.NewInviteHandler(inviteService)
	announcementHandler := handler.NewAnnouncementHandler(announcementService)
	ipBlacklistHandler := handler.NewIPBlacklistHandler(ipBlacklistService)
	healthHandler := handler.NewHealthHandler(healthService)
	adminHandler := handler.NewAdminHandler(adminService, memberService)
	memberHandler := handler.NewMemberHandler(memberService)
	watchHandler := handler.NewWatchHandler(watchService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	cardHandler := handler.NewCardHandler(cardService, cardExportService)
	memberPaymentHandler := handler.NewMemberPaymentHandler(memberPaymentService)
	vipPurchaseHandler := handler.NewVipPurchaseHandler(vipPurchaseService)
	wallpaperHandler := handler.NewWallpaperHandler(wallpaperService)
	mediaHandler := handler.NewMediaHandler(mediaService)
	imageHandler := handler.NewImageHandler(db)
	statsHandler := handler.NewStatsHandler(statsService)
	pointsHandler := handler.NewPointsHandler(pointsService)
	pointsCardHandler := handler.NewPointsCardHandler(pointsCardService)
	forumHandler := handler.NewForumHandler(forumService)
	messageHandler := handler.NewPrivateMessageHandler(messageService)
	externalCardHandler := handler.NewExternalCardHandler(externalCardService)
	goofishHandler := handler.NewGoofishHandler(goofishService)
	goofishAdminHandler := handler.NewGoofishAdminHandler(goofishService)
	goofishSignMiddleware := middleware.NewGoofishSignMiddleware(goofishService)
	paymentHandler := handler.NewPaymentHandler(alipayService)
	alipayAdminHandler := handler.NewAlipayAdminHandler(alipayService)
	cloudflareTunnelService := service.NewCloudflareTunnelService(db)
	cloudflareTunnelHandler := handler.NewCloudflareTunnelHandler(cloudflareTunnelService)

	// 初始化设置服务（提前初始化，因为公开接口也需要）
	settingService := service.NewSettingService(db)
	settingHandler := handler.NewSettingHandler(settingService)

	// API路由组
	api := r.Group("/api/v1")

	// 公开接口(无需认证)
	public := api.Group("")
	{
		// 登录限流: 5次/15分钟（基于账号，不影响其他用户）
		// 同时检测同一IP短时间内尝试多个账号的攻击行为，自动拉黑
		loginLimit := middleware.LoginRateLimit(cfg.Security.RateLimit.Login, 15*time.Minute, ipBlacklistService)
		public.POST("/user/login", loginLimit, userHandler.Login)
		public.POST("/user/send-register-code", userHandler.SendRegisterCode)
		public.POST("/user/register", userHandler.Register)
		public.POST("/user/refresh-token", userHandler.RefreshToken)
		// 忘记密码
		public.POST("/user/forgot-password", userHandler.ForgotPassword)
		public.POST("/user/reset-password", userHandler.ResetPassword)

		// 图片代理（公开访问，但需要通过URL参数传递用户ID）
		public.GET("/image/*path", imageHandler.ProxyImage)

		// 卡密续费（公开接口，允许禁用用户续费）
		public.POST("/card/renew", cardHandler.RenewByCard)

		// 网站设置（公开接口，用于前端获取网站标题等）
		public.GET("/settings/site", settingHandler.GetSiteSettingsPublic)

		// 充值链接（公开接口，用于前端获取购买链接）
		public.GET("/settings/recharge-links", settingHandler.GetRechargeLinksPublic)

		// 积分卡购买链接（公开接口）
		public.GET("/settings/points-recharge-links", settingHandler.GetPointsRechargeLinksPublic)

		// 图床设置（公开接口，用于前端获取图床地址）
		public.GET("/settings/image-host", settingHandler.GetImageHostSettingsPublic)

		// 支付宝异步通知（公开接口，无需认证）
		public.POST("/payment/alipay/notify", paymentHandler.AlipayNotify)
	}

	// 需要认证的接口
	auth := api.Group("")
	auth.Use(middleware.JWTAuth(jwtManager))
	{
		// 普通API限流: 60次/分钟
		apiLimit := middleware.RateLimit(cfg.Security.RateLimit.API, time.Minute, database.KeyAPILimit)
		auth.Use(apiLimit)

		// 用户接口
		user := auth.Group("/user")
		{
			user.GET("/info", userHandler.GetUserInfo)
			user.PUT("/update", userHandler.UpdateUserInfo)
			user.PUT("/password", userHandler.ChangePassword)
			user.POST("/avatar", userHandler.UploadAvatar)
			user.POST("/logout", userHandler.Logout)
			// 邮箱修改
			user.POST("/send-change-email-code", userHandler.SendChangeEmailCode)
			user.PUT("/email", userHandler.ChangeEmail)
			// 设备管理
			user.GET("/devices", deviceHandler.GetDevices)
			user.DELETE("/devices/:device_id", deviceHandler.RemoveDevice)
			user.DELETE("/devices", deviceHandler.RemoveAllDevices)
			// 邀请功能
			user.GET("/invite/info", inviteHandler.GetMyInviteInfo)
			user.GET("/invite/records", inviteHandler.GetMyInvites)
			user.GET("/invite/ranking", inviteHandler.GetInviteRanking)
		}

		// 会员接口
		member := auth.Group("/member")
		{
			member.GET("/info", memberHandler.GetMemberInfo)
			member.GET("/orders", memberHandler.GetMemberOrders)
			member.GET("/packages", memberPaymentHandler.GetMemberPackages)          // 获取会员套餐
			member.POST("/purchase", memberPaymentHandler.PurchaseMemberWithBalance) // 余额购买会员
		}

		// VIP购买接口
		vip := auth.Group("/vip")
		{
			vip.GET("/plans", vipPurchaseHandler.GetVipPlans)     // 获取VIP套餐列表
			vip.GET("/info", vipPurchaseHandler.GetUserVipInfo)   // 获取用户VIP信息
			vip.POST("/purchase", vipPurchaseHandler.PurchaseVip) // 购买VIP
		}

		// 支付宝支付接口
		payment := auth.Group("/payment")
		{
			payment.POST("/alipay/create", paymentHandler.CreateAlipayPayment) // 创建支付订单
			payment.GET("/order/:order_no", paymentHandler.GetOrderStatus)     // 查询订单状态
			payment.GET("/orders", paymentHandler.GetOrderList)                // 订单列表
			payment.GET("/plans", paymentHandler.GetVipPlans)                  // VIP套餐列表
			payment.GET("/member-logs", paymentHandler.GetMemberChangeLogs)    // 会员变动记录
		}

		// 媒体库接口 - 需要有效会员权限
		media := auth.Group("/media")
		media.Use(middleware.RequireMember(db)) // 会员权限检查：非会员或过期会员无法访问
		{
			media.GET("/list", mediaHandler.GetMediaDBList)                     // 获取媒体库列表
			media.GET("/db/:guid/items", mediaHandler.GetMediaDBItems)          // 获取媒体库中的媒体列表
			media.GET("/db/:guid/sum", mediaHandler.GetMediaDBSum)              // 获取媒体库统计
			media.GET("/:guid", mediaHandler.GetMediaDetail)                    // 获取媒体详情
			media.GET("/:guid/seasons", mediaHandler.GetMediaSeasons)           // 获取媒体的季列表
			media.GET("/:guid/episodes", mediaHandler.GetAllEpisodes)           // 获取剧集的所有集数
			media.GET("/season/:guid/episodes", mediaHandler.GetSeasonEpisodes) // 获取季的剧集列表
			media.GET("/search", mediaHandler.SearchMedia)                      // 搜索媒体
		}

		// 卡密接口
		card := auth.Group("/card")
		{
			// 兑换限流: 5次/小时
			redeemRateLimit := middleware.RedeemRateLimit(5, time.Hour)
			card.POST("/redeem", redeemRateLimit, cardHandler.Redeem) // 兑换卡密
			card.GET("/history", cardHandler.GetRedeemHistory)        // 兑换记录
		}

		// 观影记录接口
		watch := auth.Group("/watch")
		{
			watch.POST("/record", watchHandler.RecordWatch)
			watch.GET("/history", watchHandler.GetWatchHistory)
			watch.DELETE("/history", watchHandler.DeleteWatchHistory)
			watch.DELETE("/history/clear", watchHandler.ClearWatchHistory)
		}

		// 收藏接口
		favorite := auth.Group("/favorite")
		{
			favorite.POST("/add", watchHandler.AddFavorite)
			favorite.GET("/list", watchHandler.GetFavorites)
			favorite.DELETE("/remove", watchHandler.RemoveFavorite)
			favorite.GET("/check", watchHandler.CheckFavorite)
		}

		// 通知接口
		notification := auth.Group("/notification")
		{
			notification.GET("/list", notificationHandler.GetNotifications)
			notification.GET("/unread-count", notificationHandler.GetUnreadCount)
			notification.POST("/read", notificationHandler.MarkAsRead)
			notification.POST("/read-all", notificationHandler.MarkAllAsRead)
			notification.DELETE("/delete", notificationHandler.DeleteNotification)
		}

		// 公告接口（用户端）
		auth.GET("/announcements", announcementHandler.GetPublished)
		auth.GET("/announcement/:id", announcementHandler.GetByID)

		// 积分接口
		points := auth.Group("/points")
		{
			points.GET("/my", pointsHandler.GetMyPoints)                 // 获取我的积分
			points.POST("/sign-in", pointsHandler.SignIn)                // 签到
			points.GET("/sign-in/status", pointsHandler.GetSignInStatus) // 签到状态
			points.GET("/records", pointsHandler.GetPointsRecords)       // 积分记录
			points.GET("/exchange-rules", pointsHandler.GetExchangeRules) // 兑换规则
			points.POST("/exchange", pointsHandler.ExchangePoints)       // 积分兑换
			points.POST("/redeem", pointsCardHandler.Redeem)             // 积分卡密兑换
			points.GET("/ranking", pointsHandler.GetPointsRanking)       // 积分排行榜
			points.GET("/my-rank", pointsHandler.GetMyPointsRank)        // 我的排名
		}

		// 论坛接口
		forum := auth.Group("/forum")
		{
			forum.GET("/nodes", forumHandler.GetNodes)                                                        // 获取节点列表
			forum.POST("/topic", middleware.TopicRateLimitMiddleware(), forumHandler.CreateTopic)             // 发布话题（限流）
			forum.GET("/topics", forumHandler.GetTopicList)                                                   // 话题列表
			forum.GET("/topic/:id", forumHandler.GetTopicDetail)                                              // 话题详情
			forum.PUT("/topic/:id", forumHandler.UpdateTopic)                                                 // 更新话题
			forum.DELETE("/topic/:id", forumHandler.DeleteTopic)                                              // 删除话题
			forum.POST("/topic/:id/like", forumHandler.LikeTopic)                                             // 点赞话题
			forum.POST("/topic/:id/favorite", forumHandler.FavoriteTopic)                                     // 收藏话题
			forum.POST("/comment", middleware.CommentRateLimitMiddleware(), forumHandler.CreateComment)       // 发表评论（限流）
			forum.GET("/comments", forumHandler.GetCommentList)                                               // 评论列表
			forum.GET("/comment/:id/replies", forumHandler.GetCommentReplies)                                 // 评论回复
			forum.DELETE("/comment/:id", forumHandler.DeleteComment)                                          // 删除评论
			forum.POST("/comment/:id/like", forumHandler.LikeComment)                                         // 点赞评论
			forum.GET("/my/topics", forumHandler.GetMyTopics)                                                 // 我的话题
			forum.GET("/my/favorites", forumHandler.GetMyFavorites)                                           // 我的收藏
		}

		// 私信接口
		pm := auth.Group("/pm")
		{
			pm.POST("/send", middleware.MessageRateLimitMiddleware(), messageHandler.SendMessage) // 发送私信（限流）
			pm.GET("/conversations", messageHandler.GetConversations)                             // 会话列表
			pm.GET("/messages/:user_id", messageHandler.GetMessages)                              // 消息列表
			pm.POST("/read/:user_id", messageHandler.MarkAsRead)                                  // 标记已读
			pm.DELETE("/message/:id", messageHandler.DeleteMessage)                               // 删除消息
			pm.POST("/message/:id/recall", messageHandler.RecallMessage)                          // 撤回消息
			pm.GET("/unread-count", messageHandler.GetUnreadCount)                                // 未读数
			pm.GET("/search-users", messageHandler.SearchUsers)                                   // 搜索用户
			pm.GET("/can-send/:user_id", messageHandler.CanSendMessage)                           // 检查是否可以私信
			pm.POST("/mute/:user_id", messageHandler.MuteConversation)                            // 静音会话
		}

		// 黑名单接口
		blacklist := auth.Group("/blacklist")
		{
			blacklist.POST("/:user_id", messageHandler.BlockUser)            // 拉黑用户
			blacklist.DELETE("/:user_id", messageHandler.UnblockUser)        // 取消拉黑
			blacklist.GET("/list", messageHandler.GetBlacklist)              // 黑名单列表
			blacklist.GET("/check/:user_id", messageHandler.IsBlocked)       // 检查是否被拉黑
		}

		// 用户关注接口
		follow := auth.Group("/follow")
		{
			follow.POST("/:user_id", messageHandler.FollowUser)              // 关注/取消关注
			follow.GET("/followings", messageHandler.GetFollowings)          // 我的关注
			follow.GET("/followings/:user_id", messageHandler.GetFollowings) // 用户的关注
			follow.GET("/followers", messageHandler.GetFollowers)            // 我的粉丝
			follow.GET("/followers/:user_id", messageHandler.GetFollowers)   // 用户的粉丝
			follow.GET("/stats", messageHandler.GetFollowStats)              // 我的关注统计
			follow.GET("/stats/:user_id", messageHandler.GetFollowStats)     // 用户的关注统计
		}
	}

	// 初始化设置服务
	domainHandler := handler.NewDomainHandler(db)

	// 管理员接口(需要管理员权限)
	admin := api.Group("/admin")
	admin.Use(middleware.JWTAuth(jwtManager))
	admin.Use(middleware.RequireRole(2)) // 管理员(role >= 2)
	{
		// 系统设置
		admin.GET("/settings/email", settingHandler.GetEmailSettings)
		admin.PUT("/settings/email", settingHandler.SaveEmailSettings)
		admin.POST("/settings/email/test", settingHandler.TestEmailSettings)
		admin.GET("/settings/domain", settingHandler.GetDomainSettings)
		admin.PUT("/settings/domain", settingHandler.SaveDomainSettings)
		admin.GET("/settings/register", settingHandler.GetRegisterSettings)
		admin.PUT("/settings/register", settingHandler.SaveRegisterSettings)
		admin.GET("/settings/emby", settingHandler.GetEmbySettings)
		admin.PUT("/settings/emby", settingHandler.SaveEmbySettings)
		admin.POST("/settings/emby/test", settingHandler.TestEmbyConnection)
		admin.GET("/settings/client-whitelist", settingHandler.GetClientWhitelistSettings)
		admin.PUT("/settings/client-whitelist", settingHandler.SaveClientWhitelistSettings)
		admin.POST("/settings/client-whitelist/add", settingHandler.AddClientToWhitelist)
		admin.DELETE("/settings/client-whitelist/:name", settingHandler.RemoveClientFromWhitelist)
		admin.PUT("/settings/client-whitelist/:name/status", settingHandler.UpdateClientStatus)
		admin.GET("/settings/session-limit", settingHandler.GetSessionLimitSettings)
		admin.PUT("/settings/session-limit", settingHandler.SaveSessionLimitSettings)
		admin.GET("/settings/play-limit", settingHandler.GetPlayLimitSettings)
		admin.PUT("/settings/play-limit", settingHandler.SavePlayLimitSettings)
		admin.GET("/settings/user-cleanup", settingHandler.GetUserCleanupSettings)
		admin.PUT("/settings/user-cleanup", settingHandler.SaveUserCleanupSettings)
		admin.GET("/settings/site", settingHandler.GetSiteSettings)
		admin.PUT("/settings/site", settingHandler.SaveSiteSettings)
		admin.POST("/settings/site/logo", settingHandler.UploadLogo)
		admin.GET("/settings/recharge-links", settingHandler.GetRechargeLinksSettings)
		admin.PUT("/settings/recharge-links", settingHandler.SaveRechargeLinksSettings)
		admin.GET("/settings/points-recharge-links", settingHandler.GetPointsRechargeLinksSettings)
		admin.PUT("/settings/points-recharge-links", settingHandler.SavePointsRechargeLinksSettings)
		admin.GET("/settings/image-host", settingHandler.GetImageHostSettings)
		admin.PUT("/settings/image-host", settingHandler.SaveImageHostSettings)

		
		admin.GET("/user/list", adminHandler.GetUserList)
		admin.GET("/user/:id", adminHandler.GetUserDetail)
		admin.PUT("/user/:id/status", adminHandler.UpdateUserStatus)
		admin.PUT("/user/batch-status", adminHandler.BatchUpdateStatus)
		admin.PUT("/user/:id/reset-password", adminHandler.ResetPassword)
		admin.PUT("/user/:id/role", adminHandler.UpdateUserRole)
		admin.POST("/user/:id/set-member", adminHandler.SetMember)
		admin.PUT("/user/batch-member", adminHandler.BatchSetMember) // 批量续费
		admin.DELETE("/user/:id", adminHandler.DeleteUser)
		admin.GET("/user/:id/login-logs", adminHandler.GetLoginLogs)

		// Emby用户管理
		admin.GET("/emby/users", adminHandler.GetEmbyUsers)
		admin.GET("/emby/user/:username", adminHandler.GetEmbyUserByUsername)

		// Emby设备和会话管理
		admin.GET("/emby/sessions", adminHandler.GetEmbySessions)
		admin.GET("/emby/sessions/:username", adminHandler.GetEmbySessionsByUsername)
		admin.POST("/emby/sessions/:session_id/kill", adminHandler.KillEmbySession)
		admin.POST("/emby/enforce-session-limit", adminHandler.EnforceSessionLimit)
		admin.POST("/emby/enforce-play-limit", adminHandler.EnforcePlayLimit)
		admin.GET("/emby/session-limit-status", adminHandler.GetSessionLimitStatus)
		admin.GET("/emby/devices", adminHandler.GetEmbyDevices)
		admin.DELETE("/emby/devices/:device_id", adminHandler.DeleteEmbyDevice)
		admin.GET("/emby/user/:username/devices", adminHandler.GetEmbyDevicesByUsername)
		admin.GET("/emby/user/:username/stream-limit", adminHandler.GetEmbyUserStreamLimit)
		admin.PUT("/emby/user/:username/stream-limit", adminHandler.SetEmbyUserStreamLimit)
		admin.POST("/emby/enforce-client-whitelist", adminHandler.EnforceClientWhitelist)

		// 用户设备策略管理（使用Emby的EnableAllDevices/EnabledDevices）
		admin.GET("/emby/user/:username/device-policy", adminHandler.GetEmbyUserDevicePolicy)
		admin.PUT("/emby/user/:username/device-policy", adminHandler.SetEmbyUserDevicePolicy)
		admin.PUT("/emby/user/:username/client-policy", adminHandler.SetEmbyUserClientPolicy) // 按客户端名称设置策略
		admin.POST("/emby/user/:username/device-whitelist", adminHandler.AddDeviceToEmbyUserWhitelist)
		admin.DELETE("/emby/user/:username/device-whitelist/:device_id", adminHandler.RemoveDeviceFromEmbyUserWhitelist)
		admin.POST("/emby/user/:username/apply-global-whitelist", adminHandler.ApplyGlobalDeviceWhitelistToUser)
		admin.POST("/emby/apply-global-whitelist-all", adminHandler.ApplyGlobalDeviceWhitelistToAllUsers)

		// 用户同步
		admin.POST("/sync/user/:id/to-emby", adminHandler.SyncUserToEmby)
		admin.POST("/sync/user/:id/status", adminHandler.SyncUserStatus)
		admin.POST("/sync/user/:id/password", adminHandler.SyncUserPassword)
		admin.POST("/sync/import-emby-user", adminHandler.ImportEmbyUser)
		admin.POST("/sync/all", adminHandler.SyncAllUsers)
		admin.POST("/sync/import-all", adminHandler.ImportAllEmbyUsers)

		// 卡密管理
		admin.POST("/card/batch", cardHandler.CreateBatch)                // 批量生成卡密
		admin.GET("/card/batch/list", cardHandler.GetBatchList)           // 批次列表
		admin.GET("/card/list", cardHandler.GetCardList)                  // 卡密列表
		admin.POST("/card/:id/disable", cardHandler.DisableCard)          // 禁用卡密
		admin.POST("/card/:id/enable", cardHandler.EnableCard)            // 启用卡密
		admin.DELETE("/card/:id", cardHandler.DeleteCard)                 // 删除卡密
		admin.GET("/card/export", cardHandler.ExportCards)                // 导出卡密（旧接口，保留兼容性）
		admin.GET("/card/export/csv", cardHandler.ExportToCSV)            // 导出为CSV
		admin.GET("/card/export/excel", cardHandler.ExportToExcel)        // 导出为Excel
		admin.GET("/card/export/codes", cardHandler.ExportCodesOnly)      // 导出卡密码
		admin.GET("/card/export/report", cardHandler.GenerateUsageReport) // 生成使用报告
		admin.GET("/card/stats", cardHandler.GetCardStats)                // 卡密统计

		// 日志与统计
		admin.GET("/operation-logs", adminHandler.GetOperationLogs)
		admin.GET("/stat/user", adminHandler.GetUserStats)
		admin.GET("/stat/daily", adminHandler.GetDailyStats)

		// 公告管理
		admin.POST("/announcement", announcementHandler.Create)
		admin.GET("/announcement/list", announcementHandler.GetList)
		admin.GET("/announcement/:id", announcementHandler.GetByID)
		admin.PUT("/announcement/:id", announcementHandler.Update)
		admin.DELETE("/announcement/:id", announcementHandler.Delete)
		admin.POST("/announcement/:id/publish", announcementHandler.Publish)
		admin.POST("/announcement/:id/offline", announcementHandler.Offline)

		// IP黑名单管理
		admin.GET("/ip-blacklist", ipBlacklistHandler.GetList)
		admin.POST("/ip-blacklist", ipBlacklistHandler.Add)
		admin.DELETE("/ip-blacklist/:ip", ipBlacklistHandler.Remove)
		admin.GET("/ip-blacklist/check", ipBlacklistHandler.CheckIP)

		// 系统健康监控
		admin.GET("/system/health", healthHandler.GetHealth)
		admin.GET("/system/stats", healthHandler.GetStats)

		// 批量导入
		admin.POST("/import/users", importHandler.ImportUsers)
		admin.GET("/import/template", importHandler.GetTemplate)

		// 数据备份
		admin.POST("/backup", backupHandler.CreateBackup)
		admin.GET("/backup/list", backupHandler.GetBackupList)
		admin.GET("/backup/download/:filename", backupHandler.DownloadBackup)
		admin.DELETE("/backup/:filename", backupHandler.DeleteBackup)
		admin.POST("/backup/restore/:filename", backupHandler.RestoreBackup)

		// 邀请管理（管理后台）
		admin.GET("/invite/stats", inviteHandler.GetInviteStats)
		admin.GET("/invite/records", inviteHandler.GetInviteRecords)
		admin.POST("/invite/reward", inviteHandler.SetRewardDays)

		// 访问排行
		admin.GET("/stat/ranking", statsHandler.GetVisitRanking)

		// 积分管理
		admin.GET("/points/stats", pointsHandler.AdminGetPointsStats)
		admin.POST("/points/adjust", pointsHandler.AdminAdjustPoints)
		admin.POST("/points/gift-all", pointsHandler.AdminGiftPointsToAll) // 批量赠送积分
		admin.GET("/points/exchange-rules", pointsHandler.AdminGetExchangeRules)
		admin.POST("/points/exchange-rules", pointsHandler.AdminCreateExchangeRule)
		admin.PUT("/points/exchange-rules/:id", pointsHandler.AdminUpdateExchangeRule)
		admin.DELETE("/points/exchange-rules/:id", pointsHandler.AdminDeleteExchangeRule)

		// 积分卡密管理
		admin.POST("/points-card/batch", pointsCardHandler.AdminCreateBatch)
		admin.GET("/points-card/batch/list", pointsCardHandler.AdminGetBatchList)
		admin.DELETE("/points-card/batch/:batch_no", pointsCardHandler.AdminDeleteBatch)
		admin.GET("/points-card/list", pointsCardHandler.AdminGetCardList)
		admin.POST("/points-card/:id/disable", pointsCardHandler.AdminDisableCard)
		admin.POST("/points-card/:id/enable", pointsCardHandler.AdminEnableCard)
		admin.DELETE("/points-card/:id", pointsCardHandler.AdminDeleteCard)
		admin.GET("/points-card/stats", pointsCardHandler.AdminGetStats)
		admin.GET("/points-card/export", pointsCardHandler.AdminExportCards)

		// 积分自动赠送规则管理
		admin.GET("/points/gift-rules", pointsHandler.AdminGetGiftRules)
		admin.POST("/points/gift-rules", pointsHandler.AdminCreateGiftRule)
		admin.PUT("/points/gift-rules/:id", pointsHandler.AdminUpdateGiftRule)
		admin.DELETE("/points/gift-rules/:id", pointsHandler.AdminDeleteGiftRule)
		admin.POST("/points/gift-rules/:id/toggle", pointsHandler.AdminToggleGiftRule)
		admin.POST("/points/gift-rules/:id/execute", pointsHandler.AdminExecuteGiftRule)
		admin.GET("/points/gift-logs", pointsHandler.AdminGetGiftLogs)

		// 论坛管理
		admin.GET("/forum/nodes", forumHandler.AdminGetNodes)
		admin.POST("/forum/node", forumHandler.AdminCreateNode)
		admin.PUT("/forum/node/:id", forumHandler.AdminUpdateNode)
		admin.DELETE("/forum/node/:id", forumHandler.AdminDeleteNode)
		admin.GET("/forum/topics", forumHandler.AdminGetTopicList)
		admin.DELETE("/forum/topic/:id", forumHandler.AdminDeleteTopic)
		admin.POST("/forum/topic/:id/top", forumHandler.AdminSetTopicTop)
		admin.POST("/forum/topic/:id/recommend", forumHandler.AdminSetTopicRecommend)
		admin.DELETE("/forum/comment/:id", forumHandler.AdminDeleteComment)

		// 外部卡密API设置
		admin.GET("/settings/external-card-api", externalCardHandler.GetSettings)
		admin.PUT("/settings/external-card-api", externalCardHandler.SaveSettings)
		admin.POST("/settings/external-card-api/generate-key", externalCardHandler.GenerateAPIKey)
		admin.GET("/external-card-api/logs", externalCardHandler.GetAPILogs)

		// 闲管家虚拟货源管理
		admin.GET("/goofish/config", goofishAdminHandler.GetConfig)
		admin.POST("/goofish/config", goofishAdminHandler.SaveConfig)
		admin.GET("/goofish/goods", goofishAdminHandler.GetGoodsList)
		admin.POST("/goofish/goods", goofishAdminHandler.CreateGoods)
		admin.POST("/goofish/goods/auto-generate", goofishAdminHandler.AutoGenerateGoods)
		admin.POST("/goofish/goods/notify-all", goofishAdminHandler.NotifyAllGoodsChange)
		admin.PUT("/goofish/goods/:id", goofishAdminHandler.UpdateGoods)
		admin.DELETE("/goofish/goods/:id", goofishAdminHandler.DeleteGoods)
		admin.POST("/goofish/goods/:goods_no/notify", goofishAdminHandler.NotifyGoodsChange)
		admin.GET("/goofish/orders", goofishAdminHandler.GetOrderList)
		admin.GET("/goofish/orders/:order_no", goofishAdminHandler.GetOrderDetailAdmin)
		admin.GET("/goofish/logs", goofishAdminHandler.GetAPILogs)
		admin.POST("/goofish/logs/clean", goofishAdminHandler.CleanOldLogs)

		// 支付宝配置管理
		admin.GET("/alipay/config", alipayAdminHandler.GetConfig)
		admin.PUT("/alipay/config", alipayAdminHandler.SaveConfig)
		admin.POST("/alipay/test", alipayAdminHandler.TestConnection)
		admin.GET("/alipay/logs", alipayAdminHandler.GetLogs)
		// VIP套餐管理
		admin.GET("/alipay/plans", alipayAdminHandler.GetVipPlans)
		admin.POST("/alipay/plans", alipayAdminHandler.CreateVipPlan)
		admin.PUT("/alipay/plans/:id", alipayAdminHandler.UpdateVipPlan)
		admin.DELETE("/alipay/plans/:id", alipayAdminHandler.DeleteVipPlan)
		admin.POST("/alipay/plans/:id/toggle", alipayAdminHandler.ToggleVipPlanStatus)

		// Cloudflare 隧道管理
		admin.GET("/tunnel/status", cloudflareTunnelHandler.GetStatus)
		admin.GET("/tunnel/config", cloudflareTunnelHandler.GetConfig)
		admin.POST("/tunnel/download", cloudflareTunnelHandler.DownloadCloudflared) // 下载 cloudflared
		admin.POST("/tunnel/create", cloudflareTunnelHandler.CreateTunnel)
		admin.POST("/tunnel/start", cloudflareTunnelHandler.StartTunnel)
		admin.POST("/tunnel/stop", cloudflareTunnelHandler.StopTunnel)
		admin.POST("/tunnel/restart", cloudflareTunnelHandler.RestartTunnel)
		admin.DELETE("/tunnel", cloudflareTunnelHandler.DeleteTunnel)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 域名检查接口（公开，不需要认证）
	api.POST("/domain-check", domainHandler.CheckDomain)

	// 外部卡密API（供第三方系统调用，使用API密钥认证）
	// 同时支持GET和POST两种方式
	external := r.Group("/api/external")
	{
		external.GET("/card/fetch", externalCardHandler.FetchCard)           // 获取卡密 - GET方式
		external.POST("/card/fetch", externalCardHandler.FetchCard)          // 获取卡密 - POST方式
		external.GET("/card/fetch/:type", externalCardHandler.FetchCardByType)  // 按类型获取卡密 - GET方式
		external.POST("/card/fetch/:type", externalCardHandler.FetchCardByType) // 按类型获取卡密 - POST方式
		external.GET("/card/stock", externalCardHandler.GetStock)            // 获取库存
		
		// 咸鱼系统专用API - 简化响应格式，code字段在顶层
		// 支持咸鱼系统的所有参数: order_id, item_id, item_detail, order_amount, order_quantity, spec_name, spec_value, cookie_id, buyer_id
		external.GET("/xianyu", externalCardHandler.XianyuAutoFetchCard)     // 咸鱼专用 - 自动识别类型
		external.POST("/xianyu", externalCardHandler.XianyuAutoFetchCard)    // 咸鱼专用 - 自动识别类型
		external.GET("/xianyu/:type", externalCardHandler.XianyuFetchCard)   // 咸鱼专用 - 指定类型
		external.POST("/xianyu/:type", externalCardHandler.XianyuFetchCard)  // 咸鱼专用 - 指定类型
	}

	// 壁纸接口 (公开)
	r.GET("/api/wallpaper", wallpaperHandler.GetWallpaper)

	// 闲管家虚拟货源接口（公开，需签名验证）
	goofish := r.Group("/goofish")
	goofish.Use(goofishSignMiddleware.Verify())
	goofish.Use(goofishSignMiddleware.LogResponse())
	{
		goofish.POST("/open/info", goofishHandler.GetPlatformInfo)           // 查询平台信息
		goofish.POST("/user/info", goofishHandler.GetMerchantInfo)           // 查询商户信息
		goofish.POST("/goods/list", goofishHandler.GetGoodsList)             // 查询商品列表
		goofish.POST("/goods/detail", goofishHandler.GetGoodsDetail)         // 查询商品详情
		goofish.POST("/order/purchase/create", goofishHandler.CreateKamiOrder) // 创建卡密订单
		goofish.POST("/order/detail", goofishHandler.GetOrderDetail)         // 查询订单详情
		// 商品订阅接口
		goofish.POST("/goods/change/subscribe", goofishHandler.SubscribeGoods)           // 订阅商品变更通知
		goofish.POST("/goods/change/unsubscribe", goofishHandler.UnsubscribeGoods)       // 取消商品变更通知
		goofish.POST("/goods/change/subscribe/list", goofishHandler.GetSubscriptionList) // 查询商品订阅列表
	}

	// Emby反向代理（通过 /emby/* 访问Emby服务器）
	// 支持客户端白名单控制
	embyProxyHandler := handler.NewEmbyProxyHandler(db)
	r.Any("/emby", embyProxyHandler.ProxyEmby)
	r.Any("/emby/*path", embyProxyHandler.ProxyEmby)
	// 获取代理设置（供前端使用）
	api.GET("/emby-proxy/settings", embyProxyHandler.GetEmbyProxySettings)

	// 静态文件服务
	r.Static("/static", "./static")
	r.Static("/uploads/avatars", "./uploads/avatars")
	r.Static("/uploads/logo", "./uploads/logo")

	return r
}
