package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/liam/go_web_quick_start/scripts/generator/model"
)

// 模板文件内容
const baseServiceTemplate = `package service

import (
	"gorm.io/gorm"
)

// BaseService 提供基础CRUD操作
type BaseService struct {
	DB *gorm.DB
}

// Create 创建记录
func (s *BaseService) Create(model interface{}) error {
	return s.DB.Create(model).Error
}

// FindByID 根据ID查找记录
func (s *BaseService) FindByID(id uint, out interface{}) error {
	return s.DB.First(out, id).Error
}

// Update 更新记录
func (s *BaseService) Update(model interface{}) error {
	return s.DB.Save(model).Error
}

// Delete 删除记录
func (s *BaseService) Delete(model interface{}) error {
	return s.DB.Delete(model).Error
}

// List 列出所有记录
func (s *BaseService) List(out interface{}, page, pageSize int) error {
	offset := (page - 1) * pageSize
	return s.DB.Offset(offset).Limit(pageSize).Find(out).Error
}
`

const baseDaoTemplate = `package dao

import (
	"gorm.io/gorm"
)

// BaseDAO 提供基础数据访问操作
type BaseDAO struct {
	DB *gorm.DB
}

// Create 创建记录
func (d *BaseDAO) Create(model interface{}) error {
	return d.DB.Create(model).Error
}

// FindByID 根据ID查找记录
func (d *BaseDAO) FindByID(id uint, out interface{}) error {
	return d.DB.First(out, id).Error
}

// Update 更新记录
func (d *BaseDAO) Update(model interface{}) error {
	return d.DB.Save(model).Error
}

// Delete 删除记录
func (d *BaseDAO) Delete(model interface{}) error {
	return d.DB.Delete(model).Error
}

// List 列出所有记录
func (d *BaseDAO) List(out interface{}, page, pageSize int) error {
	offset := (page - 1) * pageSize
	return d.DB.Offset(offset).Limit(pageSize).Find(out).Error
}
`

const configLoaderTemplate = `package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"os"
)

// InitConfig 初始化配置
func InitConfig() *viper.Viper {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")

	// 检查配置文件是否存在
	if _, err := os.Stat("./config/config.yaml"); os.IsNotExist(err) {
		log.Println("警告: 配置文件不存在，将创建默认配置文件")
		
		// 确保目录存在
		if err := os.MkdirAll("./config", 0755); err != nil {
			log.Fatalf("创建配置目录失败: %v", err)
		}
		
		// 创建默认配置
		defaultConfig := []byte("# 应用配置\n" +
			"app:\n" +
			"  name: myapp\n" +
			"  port: 8080\n\n" +
			"# 数据库配置\n" +
			"database:\n" +
			"  type: mysql\n" +
			"  host: localhost\n" +
			"  port: 3306\n" +
			"  username: root\n" +
			"  password: \n" +
			"  dbname: mydb\n" +
			"  max_idle_conns: 10\n" +
			"  max_open_conns: 100\n" +
			"  log_mode: true\n\n" +
			"# Redis配置\n" +
			"redis:\n" +
			"  host: localhost\n" +
			"  port: 6379\n" +
			"  password: \n" +
			"  db: 0\n" +
			"  pool_size: 100\n")
		
		if err := os.WriteFile("./config/config.yaml", defaultConfig, 0644); err != nil {
			log.Fatalf("创建默认配置文件失败: %v", err)
		}
		log.Println("已创建默认配置文件: ./config/config.yaml")
	}

	err := v.ReadInConfig()
	if err != nil {
		log.Printf("警告: 读取配置文件失败: %v，将使用默认配置", err)
		// 设置一些默认值
		v.SetDefault("app.port", "8080")
		v.SetDefault("database.host", "localhost")
		v.SetDefault("redis.host", "localhost")
	} else {
		log.Println("配置文件加载成功")
	}

	// 监听配置文件变化
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("配置文件已更改:", e.Name)
	})

	return v
}
`

const databaseInitTemplate = `package database

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

// InitDB 初始化数据库连接
func InitDB() *gorm.DB {
	dbType := viper.GetString("database.type")
	var dialector gorm.Dialector

	switch dbType {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			viper.GetString("database.username"),
			viper.GetString("database.password"),
			viper.GetString("database.host"),
			viper.GetString("database.port"),
			viper.GetString("database.dbname"))
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
			viper.GetString("database.host"),
			viper.GetString("database.username"),
			viper.GetString("database.password"),
			viper.GetString("database.dbname"),
			viper.GetString("database.port"))
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(viper.GetString("database.dbname"))
	case "sqlserver":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
			viper.GetString("database.username"),
			viper.GetString("database.password"),
			viper.GetString("database.host"),
			viper.GetString("database.port"),
			viper.GetString("database.dbname"))
		dialector = sqlserver.Open(dsn)
	default:
		log.Fatalf("不支持的数据库类型: %s", dbType)
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取数据库连接失败: %v", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(viper.GetInt("database.max_idle_conns"))
	sqlDB.SetMaxOpenConns(viper.GetInt("database.max_open_conns"))

	return db
}
`

const redisInitTemplate = `package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"log"
	"strconv"
	"time"
)

// InitRedis 初始化Redis客户端
func InitRedis() *redis.Client {
	db, err := strconv.Atoi(viper.GetString("redis.db"))
	if err != nil {
		db = 0
		log.Printf("Redis DB转换失败，使用默认值0: %v", err)
	}

	host := viper.GetString("redis.host")
	if host == "" {
		host = "localhost"
		log.Println("未找到Redis主机配置，使用默认值localhost")
	}

	port := viper.GetString("redis.port")
	if port == "" {
		port = "6379"
		log.Println("未找到Redis端口配置，使用默认值6379")
	}

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Password:     viper.GetString("redis.password"),
		DB:           db,
		PoolSize:     viper.GetInt("redis.pool_size"),
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	_, err = client.Ping(ctx).Result()
	if err != nil {
		log.Printf("警告: 连接Redis失败: %v，但程序将继续运行\n", err)
	} else {
		log.Println("Redis连接成功")
	}

	return client
}
`

const loggerInitTemplate = `package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

// 全局日志对象
var Logger *zap.Logger

// InitLogger 初始化日志
func InitLogger() {
	// 配置编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 创建编码器
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 创建日志级别
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	// 创建core
	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
	)

	// 创建logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	defer Logger.Sync()
}
`

// 复制文件
func copyFile(src, dst string) error {
	// 读取源文件
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("读取源文件失败: %v", err)
	}

	// 写入目标文件
	return os.WriteFile(dst, content, 0644)
}

// 创建目录结构
func createProjectStructure(config model.ProjectConfig) error {
	// 创建主要目录
	directories := []string{
		"cmd",
		"config",
		"internal/api",
		"internal/dao",
		"internal/middleware",
		"internal/model",
		"internal/service",
		"pkg/cache",
		"pkg/config",
		"pkg/database",
		"pkg/logger",
		"pkg/utils",
		"pkg/wire",
		"scripts/generator",
		"scripts/generator/templates",
		"test",
	}

	for _, dir := range directories {
		path := filepath.Join(config.ProjectPath, dir)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return fmt.Errorf("创建目录 %s 失败: %v", path, err)
		}
	}

	// 先创建临时模板文件
	if err := createTemplatesFiles(config.ProjectPath); err != nil {
		return err
	}

	// 创建配置文件
	configPath := filepath.Join(config.ProjectPath, "config", "config.yaml")
	err := generateFromTemplate(configPath, "config.tmpl", config)
	if err != nil {
		return err
	}

	// 创建主程序入口
	mainPath := filepath.Join(config.ProjectPath, "cmd", "main.go")
	err = generateFromTemplate(mainPath, "main.tmpl", config)
	if err != nil {
		return err
	}

	// 创建基础服务层
	baseServicePath := filepath.Join(config.ProjectPath, "internal", "service", "base_service.go")
	err = createFileFromTemplate(baseServicePath, baseServiceTemplate, config)
	if err != nil {
		return err
	}

	// 创建基础数据访问层
	baseDaoPath := filepath.Join(config.ProjectPath, "internal", "dao", "base_dao.go")
	err = createFileFromTemplate(baseDaoPath, baseDaoTemplate, config)
	if err != nil {
		return err
	}

	// 创建配置加载器
	configLoaderPath := filepath.Join(config.ProjectPath, "pkg", "config", "config.go")
	err = createFileFromTemplate(configLoaderPath, configLoaderTemplate, config)
	if err != nil {
		return err
	}

	// 创建数据库初始化
	databaseInitPath := filepath.Join(config.ProjectPath, "pkg", "database", "database.go")
	err = createFileFromTemplate(databaseInitPath, databaseInitTemplate, config)
	if err != nil {
		return err
	}

	// 创建Redis初始化
	redisInitPath := filepath.Join(config.ProjectPath, "pkg", "cache", "redis.go")
	err = createFileFromTemplate(redisInitPath, redisInitTemplate, config)
	if err != nil {
		return err
	}

	// 创建日志初始化
	loggerInitPath := filepath.Join(config.ProjectPath, "pkg", "logger", "logger.go")
	err = createFileFromTemplate(loggerInitPath, loggerInitTemplate, config)
	if err != nil {
		return err
	}

	// 创建API路由
	apiRouterPath := filepath.Join(config.ProjectPath, "internal", "api", "router.go")
	err = generateFromTemplate(apiRouterPath, "router.tmpl", config)
	if err != nil {
		return err
	}

	// 创建用户模型文件
	userModelPath := filepath.Join(config.ProjectPath, "internal", "model", "user.go")
	err = generateFromTemplate(userModelPath, "user_model.tmpl", config)
	if err != nil {
		return err
	}

	// 创建用户DAO文件
	userDAOPath := filepath.Join(config.ProjectPath, "internal", "dao", "user_dao.go")
	err = generateFromTemplate(userDAOPath, "user_dao.tmpl", config)
	if err != nil {
		return err
	}

	// 创建用户Service文件
	userServicePath := filepath.Join(config.ProjectPath, "internal", "service", "user_service.go")
	err = generateFromTemplate(userServicePath, "user_service.tmpl", config)
	if err != nil {
		return err
	}

	// 创建用户Handler文件
	userHandlerPath := filepath.Join(config.ProjectPath, "internal", "api", "user_handler.go")
	err = generateFromTemplate(userHandlerPath, "user_handler.tmpl", config)
	if err != nil {
		return err
	}

	// 创建Wire依赖注入文件
	wirePath := filepath.Join(config.ProjectPath, "pkg", "wire", "wire.go")
	err = generateFromTemplate(wirePath, "wire.tmpl", config)
	if err != nil {
		return err
	}

	wireProviderPath := filepath.Join(config.ProjectPath, "pkg", "wire", "provider.go")
	err = generateFromTemplate(wireProviderPath, "wire_provider_base.tmpl", config)
	if err != nil {
		return err
	}

	// 创建表代码生成器
	tableGeneratorPath := filepath.Join(config.ProjectPath, "scripts", "generator", "table_generator.go")
	simpleTplContent := `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("表代码生成器")
	fmt.Println("功能待实现")
}
`
	err = os.WriteFile(tableGeneratorPath, []byte(simpleTplContent), 0644)
	if err != nil {
		return fmt.Errorf("创建表生成器文件失败: %v", err)
	}
	fmt.Println("创建了简单的表生成器文件")

	// 创建go.mod文件
	goModContent := fmt.Sprintf("module %s\n\ngo 1.20\n\nrequire (\n\tgithub.com/fsnotify/fsnotify v1.7.0\n\tgithub.com/gin-gonic/gin v1.9.1\n\tgithub.com/go-redis/redis/v8 v8.11.5\n\tgithub.com/google/wire v0.5.0\n\tgithub.com/sijms/go-ora/v2 v2.7.31\n\tgithub.com/spf13/viper v1.18.2\n\tgo.uber.org/zap v1.26.0\n\tgorm.io/driver/mysql v1.5.2\n\tgorm.io/driver/postgres v1.5.4\n\tgorm.io/driver/sqlite v1.5.4\n\tgorm.io/driver/sqlserver v1.5.2\n\tgorm.io/gorm v1.25.5\n\tgolang.org/x/crypto v0.20.0\n)\n", config.ProjectName)
	goModPath := filepath.Join(config.ProjectPath, "go.mod")
	err = os.WriteFile(goModPath, []byte(goModContent), 0644)
	if err != nil {
		return fmt.Errorf("创建go.mod文件失败: %v", err)
	}

	return nil
}

// 创建模板文件
func createTemplatesFiles(projectPath string) error {
	// 获取当前工作目录
	workDir, _ := os.Getwd()

	// 创建模板目录
	templateDir := filepath.Join(workDir, "scripts", "generator", "templates")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		return fmt.Errorf("创建模板目录失败: %v", err)
	}

	// 创建模板文件
	templates := map[string]string{
		"config.tmpl": `# 应用配置
app:
  name: {{.ProjectName}}
  port: {{.ServerPort}}

# 数据库配置
database:
  type: {{.DBType}}
  host: {{.DBHost}}
  port: {{.DBPort}}
  username: {{.DBUser}}
  password: {{.DBPassword}}
  dbname: {{.DBName}}
  max_idle_conns: 10
  max_open_conns: 100
  log_mode: true

# Redis配置
redis:
  host: {{.RedisHost}}
  port: {{.RedisPort}}
  password: {{.RedisPassword}}
  db: {{.RedisDB}}
  pool_size: 100`,
		"main.tmpl": `package main

import (
	"fmt"
	"log"
	"{{.ProjectName}}/pkg/config"
	"{{.ProjectName}}/pkg/database"
	"{{.ProjectName}}/pkg/cache"
	"{{.ProjectName}}/pkg/logger"
	"{{.ProjectName}}/internal/api"
)

func main() {
	// 初始化配置
	cfg := config.InitConfig()
	
	// 初始化日志
	logger.InitLogger()
	
	// 初始化数据库连接
	db := database.InitDB()
	
	// 初始化Redis客户端
	redis := cache.InitRedis()
	
	// 初始化路由
	router := api.InitRouter(db, redis)
	
	// 启动服务器
	port := cfg.GetString("app.port")
	log.Printf("服务启动在 http://localhost:%s", port)
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}`,
		"router.tmpl": `package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/go-redis/redis/v8"
)

// InitRouter 初始化路由
func InitRouter(db *gorm.DB, redisClient *redis.Client) *gin.Engine {
	r := gin.Default()
	
	// 中间件
	r.Use(gin.Recovery())
	
	// 健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	
	// 初始化用户相关路由
	initUserRoutes(r, db)
	
	return r
}

// 初始化用户相关路由
func initUserRoutes(r *gin.Engine, db *gorm.DB) {
	handler := NewUserHandler(db)
	
	userGroup := r.Group("/api/users")
	{
		userGroup.POST("", handler.Create)
		userGroup.GET("/:id", handler.GetByID)
		userGroup.PUT("/:id", handler.Update)
		userGroup.DELETE("/:id", handler.Delete)
		userGroup.GET("", handler.List)
	}
}`,
		"user_model.tmpl": `package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID        uint      ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
	Username  string    ` + "`json:\"username\" gorm:\"size:50;not null;unique\"`" + `
	Password  string    ` + "`json:\"password,omitempty\" gorm:\"size:100;not null\"`" + `
	Email     string    ` + "`json:\"email\" gorm:\"size:100;unique\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}`,
		"user_dao.tmpl": `package dao

import (
	"gorm.io/gorm"
	"{{.ProjectName}}/internal/model"
)

// UserDAO 用户数据访问对象
type UserDAO struct {
	BaseDAO
}

// NewUserDAO 创建用户DAO
func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		BaseDAO: BaseDAO{DB: db},
	}
}

// GetByUsername 根据用户名查找用户
func (d *UserDAO) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := d.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱查找用户
func (d *UserDAO) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := d.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID 根据ID查找用户
func (d *UserDAO) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := d.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ListUsers 获取用户列表
func (d *UserDAO) ListUsers(page, pageSize int) ([]model.User, error) {
	var users []model.User
	offset := (page - 1) * pageSize
	err := d.DB.Offset(offset).Limit(pageSize).Find(&users).Error
	return users, err
}`,
		"user_service.tmpl": `package service

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"{{.ProjectName}}/internal/dao"
	"{{.ProjectName}}/internal/model"
)

// UserService 用户服务
type UserService struct {
	BaseService
	userDAO *dao.UserDAO
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		BaseService: BaseService{DB: db},
		userDAO:     dao.NewUserDAO(db),
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(user *model.User) error {
	// 检查用户名是否已存在
	existingUser, _ := s.userDAO.GetByUsername(user.Username)
	if existingUser != nil {
		return errors.New("用户名已存在")
	}
	
	// 检查邮箱是否已存在
	if user.Email != "" {
		existingUser, _ = s.userDAO.GetByEmail(user.Email)
		if existingUser != nil {
			return errors.New("邮箱已存在")
		}
	}
	
	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	
	return s.userDAO.Create(user)
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	return s.userDAO.FindByID(id)
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(user *model.User) error {
	return s.userDAO.Update(user)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	user, err := s.userDAO.FindByID(id)
	if err != nil {
		return err
	}
	return s.userDAO.Delete(user)
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(page, pageSize int) ([]model.User, error) {
	return s.userDAO.ListUsers(page, pageSize)
}`,
		"user_handler.tmpl": `package api

import (
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	
	"{{.ProjectName}}/internal/model"
	"{{.ProjectName}}/internal/service"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{
		userService: service.NewUserService(db),
	}
}

// Create 创建用户
func (h *UserHandler) Create(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.userService.CreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// 清除密码
	user.Password = ""
	
	c.JSON(http.StatusCreated, user)
}

// GetByID 根据ID获取用户
func (h *UserHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	
	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	
	// 清除密码
	user.Password = ""
	
	c.JSON(http.StatusOK, user)
}

// Update 更新用户
func (h *UserHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	
	// 检查用户是否存在
	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	
	// 绑定请求数据
	var updateData model.User
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 更新用户
	user.Username = updateData.Username
	user.Email = updateData.Email
	if updateData.Password != "" {
		user.Password = updateData.Password
	}
	
	if err := h.userService.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// 清除密码
	user.Password = ""
	
	c.JSON(http.StatusOK, user)
}

// Delete 删除用户
func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	
	if err := h.userService.DeleteUser(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}

// List 获取用户列表
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	
	users, err := h.userService.ListUsers(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// 清除所有用户的密码
	for i := range users {
		users[i].Password = ""
	}
	
	c.JSON(http.StatusOK, users)
}`,
		"wire.tmpl": `package wire

import (
	"github.com/google/wire"
)

// InitApp 初始化应用依赖
var InitApp = wire.NewSet(
	ProvideDB,
	ProvideRedis,
	ProvideRouter,
)`,
		"wire_provider_base.tmpl": `package wire

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"{{.ProjectName}}/pkg/cache"
	"{{.ProjectName}}/pkg/database"
	"{{.ProjectName}}/internal/api"
)

// ProvideDB 提供数据库实例
func ProvideDB() *gorm.DB {
	return database.InitDB()
}

// ProvideRedis 提供Redis实例
func ProvideRedis() *redis.Client {
	return cache.InitRedis()
}

// ProvideRouter 提供路由实例
func ProvideRouter(db *gorm.DB, redis *redis.Client) *gin.Engine {
	return api.InitRouter(db, redis)
}`,
	}

	// 写入模板文件
	for name, content := range templates {
		filePath := filepath.Join(templateDir, name)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("创建模板文件: %s\n", name)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("创建模板文件 %s 失败: %v", name, err)
			}
		}
	}

	return nil
}

// 从模板文件生成
func generateFromTemplate(filePath, templateName string, data interface{}) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录 %s 失败: %v", dir, err)
	}

	// 获取当前执行文件的路径
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取执行文件路径失败: %v", err)
	}
	execDir := filepath.Dir(execPath)

	// 获取当前工作目录
	workDir, _ := os.Getwd()

	// 预先创建模板文件以确保可用
	templateDir := filepath.Join(workDir, "scripts", "generator", "templates")
	if err := os.MkdirAll(templateDir, 0755); err == nil {
		tempTemplatePath := filepath.Join(templateDir, templateName)

		// 检查模板文件是否存在，如果不存在则创建
		if _, err := os.Stat(tempTemplatePath); os.IsNotExist(err) {
			// 根据模板名创建基本内容
			var templateContent string

			if templateName == "config.tmpl" {
				templateContent = `# 应用配置
app:
  name: {{.ProjectName}}
  port: {{.ServerPort}}

# 数据库配置
database:
  type: {{.DBType}}
  host: {{.DBHost}}
  port: {{.DBPort}}
  username: {{.DBUser}}
  password: {{.DBPassword}}
  dbname: {{.DBName}}
  max_idle_conns: 10
  max_open_conns: 100
  log_mode: true

# Redis配置
redis:
  host: {{.RedisHost}}
  port: {{.RedisPort}}
  password: {{.RedisPassword}}
  db: {{.RedisDB}}
  pool_size: 100`
			} else if templateName == "main.tmpl" {
				templateContent = `package main

import (
	"fmt"
	"log"
	"{{.ProjectName}}/pkg/config"
	"{{.ProjectName}}/pkg/database"
	"{{.ProjectName}}/pkg/cache"
	"{{.ProjectName}}/pkg/logger"
	"{{.ProjectName}}/internal/api"
)

func main() {
	// 初始化配置
	cfg := config.InitConfig()
	
	// 初始化日志
	logger.InitLogger()
	
	// 初始化数据库连接
	db := database.InitDB()
	
	// 初始化Redis客户端
	redis := cache.InitRedis()
	
	// 初始化路由
	router := api.InitRouter(db, redis)
	
	// 启动服务器
	port := cfg.GetString("app.port")
	log.Printf("服务启动在 http://localhost:%s", port)
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
`
			} else if templateName == "router.tmpl" {
				templateContent = `package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/go-redis/redis/v8"
)

// InitRouter 初始化路由
func InitRouter(db *gorm.DB, redisClient *redis.Client) *gin.Engine {
	r := gin.Default()
	
	// 中间件
	r.Use(gin.Recovery())
	
	// 健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	
	// 初始化用户相关路由
	initUserRoutes(r, db)
	
	return r
}

// 初始化用户相关路由
func initUserRoutes(r *gin.Engine, db *gorm.DB) {
	handler := NewUserHandler(db)
	
	userGroup := r.Group("/api/users")
	{
		userGroup.POST("", handler.Create)
		userGroup.GET("/:id", handler.GetByID)
		userGroup.PUT("/:id", handler.Update)
		userGroup.DELETE("/:id", handler.Delete)
		userGroup.GET("", handler.List)
	}
}`
			} else if templateName == "user_model.tmpl" {
				templateContent = `package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID        uint      ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
	Username  string    ` + "`json:\"username\" gorm:\"size:50;not null;unique\"`" + `
	Password  string    ` + "`json:\"password,omitempty\" gorm:\"size:100;not null\"`" + `
	Email     string    ` + "`json:\"email\" gorm:\"size:100;unique\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}`
			} else if templateName == "user_dao.tmpl" {
				templateContent = `package dao

import (
	"gorm.io/gorm"
	"{{.ProjectName}}/internal/model"
)

// UserDAO 用户数据访问对象
type UserDAO struct {
	BaseDAO
}

// NewUserDAO 创建用户DAO
func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		BaseDAO: BaseDAO{DB: db},
	}
}

// GetByUsername 根据用户名查找用户
func (d *UserDAO) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := d.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱查找用户
func (d *UserDAO) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := d.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID 根据ID查找用户
func (d *UserDAO) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := d.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ListUsers 获取用户列表
func (d *UserDAO) ListUsers(page, pageSize int) ([]model.User, error) {
	var users []model.User
	offset := (page - 1) * pageSize
	err := d.DB.Offset(offset).Limit(pageSize).Find(&users).Error
	return users, err
}`
			} else if templateName == "user_service.tmpl" {
				templateContent = `package service

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"{{.ProjectName}}/internal/dao"
	"{{.ProjectName}}/internal/model"
)

// UserService 用户服务
type UserService struct {
	BaseService
	userDAO *dao.UserDAO
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		BaseService: BaseService{DB: db},
		userDAO:     dao.NewUserDAO(db),
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(user *model.User) error {
	// 检查用户名是否已存在
	existingUser, _ := s.userDAO.GetByUsername(user.Username)
	if existingUser != nil {
		return errors.New("用户名已存在")
	}
	
	// 检查邮箱是否已存在
	if user.Email != "" {
		existingUser, _ = s.userDAO.GetByEmail(user.Email)
		if existingUser != nil {
			return errors.New("邮箱已存在")
		}
	}
	
	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	
	return s.userDAO.Create(user)
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	return s.userDAO.FindByID(id)
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(user *model.User) error {
	return s.userDAO.Update(user)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	user, err := s.userDAO.FindByID(id)
	if err != nil {
		return err
	}
	return s.userDAO.Delete(user)
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(page, pageSize int) ([]model.User, error) {
	return s.userDAO.ListUsers(page, pageSize)
}`
			} else if templateName == "user_handler.tmpl" {
				templateContent = `package api

import (
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	
	"{{.ProjectName}}/internal/model"
	"{{.ProjectName}}/internal/service"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{
		userService: service.NewUserService(db),
	}
}

// Create 创建用户
func (h *UserHandler) Create(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.userService.CreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// 清除密码
	user.Password = ""
	
	c.JSON(http.StatusCreated, user)
}

// GetByID 根据ID获取用户
func (h *UserHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	
	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	
	// 清除密码
	user.Password = ""
	
	c.JSON(http.StatusOK, user)
}

// Update 更新用户
func (h *UserHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	
	// 检查用户是否存在
	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	
	// 绑定请求数据
	var updateData model.User
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 更新用户
	user.Username = updateData.Username
	user.Email = updateData.Email
	if updateData.Password != "" {
		user.Password = updateData.Password
	}
	
	if err := h.userService.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// 清除密码
	user.Password = ""
	
	c.JSON(http.StatusOK, user)
}

// Delete 删除用户
func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	
	if err := h.userService.DeleteUser(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}

// List 获取用户列表
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	
	users, err := h.userService.ListUsers(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// 清除所有用户的密码
	for i := range users {
		users[i].Password = ""
	}
	
	c.JSON(http.StatusOK, users)
}`
			} else if templateName == "wire.tmpl" {
				templateContent = `package wire

import (
	"github.com/google/wire"
)

// InitApp 初始化应用依赖
var InitApp = wire.NewSet(
	ProvideDB,
	ProvideRedis,
	ProvideRouter,
)`
			} else if templateName == "wire_provider_base.tmpl" {
				templateContent = `package wire

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"{{.ProjectName}}/pkg/cache"
	"{{.ProjectName}}/pkg/database"
	"{{.ProjectName}}/internal/api"
)

// ProvideDB 提供数据库实例
func ProvideDB() *gorm.DB {
	return database.InitDB()
}

// ProvideRedis 提供Redis实例
func ProvideRedis() *redis.Client {
	return cache.InitRedis()
}

// ProvideRouter 提供路由实例
func ProvideRouter(db *gorm.DB, redis *redis.Client) *gin.Engine {
	return api.InitRouter(db, redis)
}`
			} else {
				// 其他模板采用简单内容
				templateContent = fmt.Sprintf("// %s 模板文件\n// 用于 %s\n", templateName, templateName)
			}

			// 写入模板文件
			err := os.WriteFile(tempTemplatePath, []byte(templateContent), 0644)
			if err == nil {
				fmt.Printf("创建临时模板文件: %s\n", tempTemplatePath)
			}
		}
	}

	// 读取模板文件 - 尝试多个可能的路径
	possiblePaths := []string{
		filepath.Join(workDir, "scripts", "generator", "templates", templateName),
		filepath.Join(workDir, templateName),
		templateName,
		filepath.Join("scripts", "generator", "templates", templateName),
		filepath.Join("templates", templateName),
		filepath.Join(execDir, "scripts", "generator", "templates", templateName),
		filepath.Join(execDir, "..", "scripts", "generator", "templates", templateName),
	}

	var templateContent []byte
	var readErr error

	for _, path := range possiblePaths {
		fmt.Printf("尝试读取模板文件: %s\n", path)
		templateContent, readErr = os.ReadFile(path)
		if readErr == nil {
			fmt.Printf("成功读取模板文件: %s\n", path)
			break
		}
	}

	if readErr != nil {
		return fmt.Errorf("读取模板文件 %s 失败: %v\n尝试的路径: %v", templateName, readErr, possiblePaths)
	}

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件 %s 失败: %v", filePath, err)
	}
	defer file.Close()

	// 解析模板
	tmpl, err := template.New(filepath.Base(templateName)).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	// 执行模板
	err = tmpl.Execute(file, data)
	if err != nil {
		return fmt.Errorf("执行模板失败: %v", err)
	}

	return nil
}

// 从模板创建文件（传入模板字符串）
func createFileFromTemplate(filePath, templateContent string, data interface{}) error {
	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件 %s 失败: %v", filePath, err)
	}
	defer file.Close()

	// 解析模板
	tmpl, err := template.New(filepath.Base(filePath)).Parse(templateContent)
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	// 执行模板
	err = tmpl.Execute(file, data)
	if err != nil {
		return fmt.Errorf("执行模板失败: %v", err)
	}

	return nil
}

// 获取用户输入
func getUserInput(prompt string, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}
	return input
}

// 创建项目的主函数
func CreateProject() {
	// 创建项目
	fmt.Println("=== Go Web Quick Start 项目生成器 ===")
	fmt.Println("请输入以下信息来生成您的项目：")

	// 获取项目信息
	config := model.ProjectConfig{}

	// 项目名称
	config.ProjectName = getUserInput("项目名称（Go模块名）", "github.com/yourusername/myproject")

	// 确保项目名称格式正确
	if !strings.Contains(config.ProjectName, "/") {
		fmt.Println("警告: 项目名称应该是完整的模块路径，例如 github.com/username/project")
		fmt.Println("将自动修正为标准格式...")
		config.ProjectName = "github.com/" + config.ProjectName
	}

	// 项目路径
	defaultPath, _ := os.Getwd()
	config.ProjectPath = getUserInput("项目路径", defaultPath)

	// 数据库类型
	config.DBType = getUserInput("数据库类型 (mysql, postgres, sqlite, sqlserver, oracle)", "mysql")

	// 数据库连接信息
	config.DBHost = getUserInput("数据库主机", "localhost")
	config.DBPort = getUserInput("数据库端口", "3306")
	config.DBUser = getUserInput("数据库用户名", "root")
	config.DBPassword = getUserInput("数据库密码", "")
	config.DBName = getUserInput("数据库名", "mydb")

	// Redis连接信息
	config.RedisHost = getUserInput("Redis主机", "localhost")
	config.RedisPort = getUserInput("Redis端口", "6379")
	config.RedisPassword = getUserInput("Redis密码", "")
	config.RedisDB = getUserInput("Redis数据库", "0")

	// 服务器端口
	config.ServerPort = getUserInput("服务器端口", "8080")

	// 创建项目
	fmt.Println("\n正在生成项目...")
	err := createProjectStructure(config)
	if err != nil {
		fmt.Printf("生成项目失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n项目生成成功！")
	fmt.Printf("项目位置: %s\n", config.ProjectPath)
	fmt.Println("\n要启动项目，请执行以下命令：")
	fmt.Printf("cd %s\n", config.ProjectPath)
	fmt.Println("go mod tidy")
	fmt.Println("go run cmd/main.go")
}
