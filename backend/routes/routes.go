package routes

import (
	"go-yzs/handlers"
	"go-yzs/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) {
	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")

	// Public routes
	api.POST("/login", handlers.Login)

	// Authenticated routes
	auth := api.Group("")
	auth.Use(middleware.AuthRequired())
	{
		auth.POST("/logout", handlers.Logout)
		auth.GET("/me", handlers.GetCurrentUser)
		auth.POST("/me/start", handlers.IncrementStart)
		auth.POST("/me/skip", handlers.IncrementSkip)
		auth.GET("/me/daily-stats", handlers.GetMyDailyStats)
		auth.GET("/me/inspect-stats", handlers.GetMyInspectStats)
		auth.POST("/me/handled-goods", handlers.SaveHandledGoods)
		auth.GET("/me/handled-goods", handlers.ListMyHandledGoods)
		auth.POST("/me/password", handlers.ChangePassword)

		// Trade abnormal data
		auth.GET("/trades", handlers.ListTradeAbnormal)
		auth.GET("/trades/hourly-stats", handlers.GetHourlyStats)
		auth.GET("/trades/my-handled", handlers.ListMyHandled) // 固定路径必须在 /:id 之前
		auth.GET("/trades/random-unhandled", handlers.GetRandomUnhandled)
		auth.GET("/trades/random-uninspected", handlers.GetRandomUninspected)
		auth.POST("/trades/:id/handle", handlers.HandleTrade)
		auth.POST("/trades/:id/pend", handlers.PendTrade)
		auth.POST("/trades/:id/lock", handlers.LockTrade)
		auth.POST("/trades/:id/unlock", handlers.UnlockTrade)
		auth.GET("/trades/:id/check", handlers.CheckTradeStatus)
		auth.POST("/trades/:id/submit", handlers.SubmitTrade)
		auth.POST("/trades/:id/inspect", handlers.InspectTrade)
		auth.GET("/trades/:id/detail", handlers.GetTradeDetail)
		auth.GET("/trades/:id/branch-products", handlers.QueryBranchProducts)
		auth.POST("/trades/:id/product-price", handlers.QueryProductPrice)
		auth.GET("/stats", handlers.GetStats)
		auth.GET("/stats/operators", handlers.GetOperatorStats)
		auth.GET("/stats/operator-records", handlers.GetOperatorRecords)
		auth.GET("/stats/daily", handlers.GetDailyStats)
		auth.GET("/stats/operator-range", handlers.GetOperatorRangeStats)
		auth.GET("/stats/inspect-export", handlers.GetInspectExport)
		auth.GET("/stats/inspectors", handlers.GetInspectorStats)
		auth.GET("/stats/inspector-range", handlers.GetInspectorRangeStats)

		// Quality review
		auth.GET("/reviews", handlers.ListReviews)
		auth.POST("/reviews/:id/approve", handlers.ApproveReview)
		auth.POST("/reviews/:id/remark", handlers.AddReviewRemark)

		// Goods catalog
		auth.GET("/goods", handlers.ListGoods)

		// Favorite goods
		auth.POST("/favorites", handlers.AddFavoriteGoods)
		auth.DELETE("/favorites/:goodsId", handlers.RemoveFavoriteGoods)
		auth.GET("/favorites", handlers.ListFavoriteGoods)
		auth.GET("/favorites/check", handlers.CheckFavoriteGoods)

		// User management (admin only)
		admin := auth.Group("/users")
		admin.Use(middleware.AdminRequired())
		{
			admin.GET("", handlers.ListUsers)
			admin.POST("", handlers.CreateUser)
			admin.PUT("/:id", handlers.UpdateUser)
			admin.DELETE("/:id", handlers.DeleteUser)
		}

		// Team management (admin only)
		teamAdmin := auth.Group("/teams")
		teamAdmin.Use(middleware.AdminRequired())
		{
			teamAdmin.GET("", handlers.ListTeams)
			teamAdmin.POST("", handlers.CreateTeam)
			teamAdmin.PUT("/:id", handlers.UpdateTeam)
			teamAdmin.DELETE("/:id", handlers.DeleteTeam)
		}
	}
}
