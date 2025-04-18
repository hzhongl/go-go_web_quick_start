package model

// ModelConfig 存储用户输入的模型配置信息
type ModelConfig struct {
	ModuleName    string
	TableName     string
	ModelName     string
	Fields        []Field
	ProjectImport string
	ID            string // ID类型
	DBType        string // 数据库类型
}

// ProjectConfig 存储用户输入的项目配置信息
type ProjectConfig struct {
	ProjectName   string
	ProjectPath   string
	DBType        string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       string
	ServerPort    string
}

// TableConfig 表配置
type TableConfig struct {
	ModuleName    string
	TableName     string
	ModelName     string
	Fields        []Field
	ProjectImport string
}

// Field 字段定义
type Field struct {
	Name    string
	Type    string
	Tag     string
	Comment string
}
