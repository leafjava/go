package main

import (
	"errors"
	"fmt"
)

// 方式1：返回 error
func connectDatabase(host string) error {
	if host == "" {
		return errors.New("host is empty")
	}

	fmt.Println("连接数据库:", host)
	return nil
}

// 方式2：自定义错误类型
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("字段 %s 验证失败: %s", e.Field, e.Message)
}

func validateUser(name string, age int) error {
	if name == "" {
		return &ValidationError{
			Field:   "name",
			Message: "用户名不能为空",
		}
	}

	if age < 18 {
		return &ValidationError{
			Field:   "age",
			Message: "年龄必须大于18",
		}
	}

	return nil
}

// 方式3：包装错误（Go 1.13+）
func processTransaction(txHash string) error {
	if txHash == "" {
		return errors.New("交易哈希为空")
	}

	err := errors.New("网络超时")
	if err != nil {
		return fmt.Errorf("交易处理失败 %s: %w", txHash, err)
	}

	return nil

}

//ve, ok := err.(*ValidationError)
//err.(*ValidationError) - 尝试将 err 转换为 *ValidationError 类型
//ve - 如果转换成功，这是转换后的值
//ok - 布尔值，true 表示转换成功，false 表示失败

func main() {
	if err := connectDatabase("localhost"); err != nil {
		fmt.Println(err)
		return
	}

	// 类型断言检查错误类型
	if err := validateUser("", 20); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			fmt.Printf("验证错误 - 字段: %s, 消息: %s\n", ve.Field, ve.Message)
		} else {
			fmt.Println("未知错误:", err)
		}
	}

	// 错误包装
	if err := processTransaction("0xabc123"); err != nil {
		fmt.Println("错误:", err)
	}
}
