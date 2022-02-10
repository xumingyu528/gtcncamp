package main

/*
课后练习 1.1
编写一个小程序：
给定一个字符串数组
[“I”,“am”,“stupid”,“and”,“weak”]
用 for 循环遍历该数组并修改为
[“I”,“am”,“smart”,“and”,“strong”]
*/

import "fmt"

func main() {
	arrStr := []string{"I", "am", "stupid", "and", "weak"}

	fmt.Print("before join camp, ")
	fmt.Println(arrStr)
	for i := range arrStr {
		if arrStr[i] == "stupid" {
			arrStr[i] = "smart"
		}

		if arrStr[i] == "weak" {
			arrStr[i] = "strong"
		}
	}
	fmt.Print("but after that, ")
	fmt.Println(arrStr)
}
