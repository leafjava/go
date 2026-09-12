package main

import (
	"errors"
	"fmt"
)

func connectDatabase(host string) error {
	if host == "" {
		return errors.New("主机地址不能为空")
	}

	fmt.Println("连接数据库:", host)
	return nil
}

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
			Message: "年龄必须大于18岁",
		}
	}

	return nil
}

func processTransaction(txHash string) error {
	if txHash == "" {
		return errors.New("交易哈希为空")
	}

	err := errors.New("网络超时")
	if err != nil {
		return fmt.Errorf("处理交易失败 %s: %w", txHash, err)
	}

	return nil
}

func main() {
	if err := connectDatabase("localhost"); err != nil {
		fmt.Println("错误:", err)
		return
	}

	if err := validateUser("", 20); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			fmt.Printf("验证错误 - 字段: %s, 消息: %s\n", ve.Field, ve.Message)
		} else {
			fmt.Println("未知错误:", err)
		}
	}

	if err := processTransaction("0xabc123"); err != nil {
		fmt.Println("错误:", err)
	}
}
