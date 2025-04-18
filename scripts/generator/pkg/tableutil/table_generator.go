package tableutil

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

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

// Field 字段定义
type Field struct {
	Name    string
	Type    string
	Tag     string
	Comment string
}

// 获取用户输入
func GetUserInput(prompt string, defaultValue string) string {
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

// 获取布尔值输入
func GetBoolInput(prompt string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [%s]: ", prompt, defaultStr)

	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	if input == "" {
		return defaultValue
	}

	return input == "y" || input == "yes" || input == "true"
}

// 根据数据库类型获取ID类型
func GetIDType(dbType string) string {
	switch dbType {
	case "mysql", "postgres", "sqlserver":
		return "uint"
	case "oracle":
		return "int64"
	case "sqlite":
		return "int"
	default:
		return "uint"
	}
}

// 生成文件从模板
func GenerateFileFromTemplate(filePath, templatePath string, data interface{}) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 读取模板文件
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("读取模板文件失败: %v", err)
	}

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 解析模板
	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	// 执行模板
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("执行模板失败: %v", err)
	}

	return nil
}

// 更新wire provider文件
func UpdateWireProvider(filePath string, modelName, tableName, moduleName string) error {
	// 读取现有内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	// 读取wire_provider.tmpl模板
	templatePath := filepath.Join("scripts", "generator", "templates", "wire_provider.tmpl")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("读取模板文件失败: %v", err)
	}

	// 找到合适的位置插入新内容（在ProviderSet定义之前）
	lines := strings.Split(string(content), "\n")
	var newContent strings.Builder

	providerSetIndex := -1
	for i, line := range lines {
		if strings.Contains(line, "var ProviderSet = wire.NewSet(") {
			providerSetIndex = i
			break
		}
	}

	if providerSetIndex == -1 {
		return fmt.Errorf("未找到ProviderSet定义")
	}

	// 解析模板
	tmpl, err := template.New("wire").Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("解析Wire模板失败: %v", err)
	}

	// 准备数据
	data := struct {
		ModelName  string
		TableName  string
		ModuleName string
	}{
		ModelName:  modelName,
		TableName:  tableName,
		ModuleName: moduleName,
	}

	// 执行模板
	var wireContent strings.Builder
	err = tmpl.Execute(&wireContent, data)
	if err != nil {
		return fmt.Errorf("生成Wire内容失败: %v", err)
	}

	// 添加新的依赖注入集合
	for i, line := range lines {
		newContent.WriteString(line)
		newContent.WriteString("\n")

		if i == providerSetIndex-1 {
			newContent.WriteString(wireContent.String())
			newContent.WriteString("\n")
		}
	}

	// 写回文件
	return os.WriteFile(filePath, []byte(newContent.String()), 0644)
}

// 更新ProviderSet定义
func UpdateProviderSet(filePath string, newSet string) error {
	// 读取现有内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	// 找到ProviderSet定义
	lines := strings.Split(string(content), "\n")
	var newContent strings.Builder

	for i, line := range lines {
		if strings.Contains(line, "var ProviderSet = wire.NewSet(") {
			// 找到下一个")"
			closingIndex := -1
			for j := i + 1; j < len(lines); j++ {
				if strings.Contains(lines[j], ")") {
					closingIndex = j
					break
				}
			}

			if closingIndex != -1 {
				// 插入新的Set
				for k := 0; k <= i; k++ {
					newContent.WriteString(lines[k])
					newContent.WriteString("\n")
				}

				// 检查是否已经有其他Set
				if !strings.Contains(lines[i+1], "//") {
					newContent.WriteString("\t" + newSet + ",\n")
				} else {
					lines[closingIndex-1] = "\t" + newSet + ","
				}

				for k := i + 1; k < len(lines); k++ {
					newContent.WriteString(lines[k])
					newContent.WriteString("\n")
				}

				// 写回文件
				return os.WriteFile(filePath, []byte(newContent.String()), 0644)
			}
		}
	}

	return fmt.Errorf("未找到合适的位置插入新Set")
}

// 表结构生成器主函数
func GenerateTable() {
	fmt.Println("=== 数据表代码生成器 ===")
	fmt.Println("请输入以下信息来生成代码：")

	// 获取项目导入路径
	projectImport := GetUserInput("项目导入路径", "github.com/yourusername/myproject")

	// 获取数据库类型
	dbType := GetUserInput("数据库类型 (mysql, postgres, sqlite, sqlserver, oracle)", "mysql")

	// 表名
	tableName := GetUserInput("表名", "users")

	// 获取字段信息
	fmt.Println("\n请输入字段信息（每行一个字段，格式：字段名 类型 标签 注释，输入空行结束）：")
	fmt.Println("例如：name string `gorm:\"type:varchar(100)\" json:\"name\"` 名称")

	var fields []Field
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		scanner.Scan()
		line := scanner.Text()
		if line == "" {
			break
		}

		parts := strings.SplitN(line, " ", 4)
		field := Field{
			Name: parts[0],
			Type: parts[1],
		}

		if len(parts) > 2 {
			field.Tag = parts[2]
		}

		if len(parts) > 3 {
			field.Comment = parts[3]
		}

		fields = append(fields, field)
	}

	// 获取模型名称（首字母大写）
	defaultModelName := strings.ToUpper(tableName[:1]) + tableName[1:]
	if strings.HasSuffix(defaultModelName, "s") {
		defaultModelName = defaultModelName[:len(defaultModelName)-1] // 去掉复数形式
	}
	modelName := GetUserInput("模型名称", defaultModelName)

	// 获取模块名（首字母小写）
	defaultModuleName := strings.ToLower(modelName[:1]) + modelName[1:]
	moduleName := GetUserInput("模块名称", defaultModuleName)

	// 获取ID类型
	idType := GetIDType(dbType)

	// 创建配置
	config := ModelConfig{
		ModuleName:    moduleName,
		TableName:     tableName,
		ModelName:     modelName,
		Fields:        fields,
		ProjectImport: projectImport,
		ID:            idType,
		DBType:        dbType,
	}

	// 确认项目根目录
	projectRoot := GetUserInput("项目根目录", ".")

	// 模板目录
	templatesDir := filepath.Join("scripts", "generator", "templates")

	// 生成模型文件
	modelPath := filepath.Join(projectRoot, "internal", "model", strings.ToLower(moduleName)+".go")
	err := GenerateFileFromTemplate(modelPath, filepath.Join(templatesDir, "model.tmpl"), config)
	if err != nil {
		fmt.Printf("生成模型文件失败: %v\n", err)
		os.Exit(1)
	}

	// 生成DAO文件
	daoPath := filepath.Join(projectRoot, "internal", "dao", strings.ToLower(moduleName)+"_dao.go")
	err = GenerateFileFromTemplate(daoPath, filepath.Join(templatesDir, "dao.tmpl"), config)
	if err != nil {
		fmt.Printf("生成DAO文件失败: %v\n", err)
		os.Exit(1)
	}

	// 生成Service文件
	servicePath := filepath.Join(projectRoot, "internal", "service", strings.ToLower(moduleName)+"_service.go")
	err = GenerateFileFromTemplate(servicePath, filepath.Join(templatesDir, "service.tmpl"), config)
	if err != nil {
		fmt.Printf("生成Service文件失败: %v\n", err)
		os.Exit(1)
	}

	// 生成Handler文件
	handlerPath := filepath.Join(projectRoot, "internal", "api", strings.ToLower(moduleName)+"_handler.go")
	err = GenerateFileFromTemplate(handlerPath, filepath.Join(templatesDir, "handler.tmpl"), config)
	if err != nil {
		fmt.Printf("生成Handler文件失败: %v\n", err)
		os.Exit(1)
	}

	// 更新Wire Provider
	wireProviderPath := filepath.Join(projectRoot, "pkg", "wire", "provider.go")

	// 更新Wire Provider文件
	err = UpdateWireProvider(wireProviderPath, modelName, tableName, moduleName)
	if err != nil {
		// 尝试更新ProviderSet
		updateErr := UpdateProviderSet(wireProviderPath, modelName+"Set")
		if updateErr != nil {
			fmt.Printf("更新Wire Provider失败: %v\n", updateErr)
			os.Exit(1)
		}
	}

	fmt.Println("\n代码生成成功！")
	fmt.Printf("模型文件: %s\n", modelPath)
	fmt.Printf("DAO文件: %s\n", daoPath)
	fmt.Printf("Service文件: %s\n", servicePath)
	fmt.Printf("Handler文件: %s\n", handlerPath)
	fmt.Println("\n已更新Wire依赖注入")
}
