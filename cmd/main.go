package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/geolav/QEasyApp/internal/controller/http/middleware"
	httpQuiz "github.com/geolav/QEasyApp/internal/controller/http/quiz"
	httpSession "github.com/geolav/QEasyApp/internal/controller/http/session"
	httpUser "github.com/geolav/QEasyApp/internal/controller/http/user"
	wsHub "github.com/geolav/QEasyApp/internal/controller/websocket"
	repoQuiz "github.com/geolav/QEasyApp/internal/repository/quiz"
	repoSession "github.com/geolav/QEasyApp/internal/repository/session"
	repoUser "github.com/geolav/QEasyApp/internal/repository/user"
	ucQuiz "github.com/geolav/QEasyApp/internal/usecase/quiz"
	ucSession "github.com/geolav/QEasyApp/internal/usecase/session"
	ucUser "github.com/geolav/QEasyApp/internal/usecase/user"
	"github.com/gin-gonic/gin"

	"github.com/geolav/QEasyApp/config"
	"github.com/geolav/QEasyApp/internal/repository"
)

func main() {
	cfg := config.LoadConfig()
	db, err := repository.NewPostgresPool(
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	userRepo := repoUser.New(db)
	quizRepo := repoQuiz.New(db)
	sessionRepo := repoSession.New(db)

	userUC := ucUser.New(userRepo, cfg.JWTSecret)
	quizUC := ucQuiz.New(quizRepo)
	sessionUC := ucSession.New(sessionRepo, quizRepo)

	userHandler := httpUser.New(userUC)
	quizHandler := httpQuiz.New(quizUC)
	sessionHandler := httpSession.New(sessionUC)

	// ========== ДОБАВЬТЕ ЭТУ ПРОВЕРКУ ==========
	// Определяем путь к корню проекта
	_, currentFile, _, _ := runtime.Caller(0)
	// currentFile = /home/yegor/GolandProjects/QEasyApp/cmd/main.go
	projectRoot := filepath.Dir(filepath.Dir(currentFile))
	// projectRoot = /home/yegor/GolandProjects/QEasyApp

	frontendPath := filepath.Join(projectRoot, "frontend")
	log.Printf("Looking for frontend at: %s", frontendPath)
	// ============================================

	hub := wsHub.NewHub()
	go hub.Run()

	wsHandler := wsHub.NewHandler(hub, sessionUC, quizUC)

	r := gin.Default()

	public := r.Group("/api")
	{
		public.POST("/register", userHandler.Register)
		public.POST("/login", userHandler.Login)
	}
	protected := r.Group("/api")
	protected.Use(middleware.Auth(cfg.JWTSecret))
	{
		protected.GET("/profile", userHandler.GetProfile)

		protected.POST("/quizzes", quizHandler.CreateQuiz)
		protected.GET("/quizzes", quizHandler.GetMyQuizzes)
		protected.GET("/quizzes/:quiz_id", quizHandler.GetQuiz)
		protected.POST("/quizzes/:quiz_id/questions", quizHandler.AddQuestion)
		protected.POST("/quizzes/:quiz_id/publish", quizHandler.PublishQuiz)

		protected.POST("/sessions", sessionHandler.CreateSession)
		protected.POST("/sessions/join", sessionHandler.JoinSession)
		protected.POST("/sessions/:session_id/start", sessionHandler.StartSession)
		protected.POST("/sessions/:session_id/next", sessionHandler.NextQuestion)
		protected.POST("/sessions/:session_id/finish", sessionHandler.FinishSession)
		protected.GET("/sessions/:session_id/leaderboard", sessionHandler.GetLeaderboard)

		protected.GET("/ws/:session_id", wsHandler.Connect)
	}

	//// ========== ФРОНТЕНД ==========
	//r.Static("/static", filepath.Join(frontendPath))
	//r.StaticFile("/", filepath.Join(frontendPath, "index.html"))
	//
	//// Опционально: для favicon
	//if _, err := os.Stat(filepath.Join(frontendPath, "favicon.ico")); err == nil {
	//	r.StaticFile("/favicon.ico", filepath.Join(frontendPath, "favicon.ico"))
	//}
	//
	//r.NoRoute(func(c *gin.Context) {
	//	c.File(filepath.Join(frontendPath, "index.html"))
	//})
	//// ===============================

	// ========== ФРОНТЕНД ==========
	// Раздача статических файлов из папки frontend
	r.Static("/static", "./frontend")
	r.StaticFile("/", "./frontend/index.html")

	// Опционально: для favicon
	if _, err := os.Stat("./frontend/favicon.ico"); err == nil {
		r.StaticFile("/favicon.ico", "./frontend/favicon.ico")
	}

	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/index.html")
	})
	// ===============================

	log.Printf("server started on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
