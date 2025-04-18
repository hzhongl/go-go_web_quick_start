package main

import (
	"fmt"
	"os"
	"os/exec"
)

// 从project_generator.go导入CreateProject函数
//go:generate go run project_generator.go

// 这里声明CreateProject函数的引用
// CreateProject在project_generator.go中定义
// 由于两个文件在同一个包中，所以可以直接引用
// 此注释仅为文档目的

func main() {
	args := os.Args
	if len(args) > 1 && args[1] == "create-table" {

		cmd := exec.Command("go", "run", "./cmd/table_generator/main.go")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			fmt.Printf("运行表代码生成器失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		// 调用项目生成函数
		CreateProject()
	}
}
