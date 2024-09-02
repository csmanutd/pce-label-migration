package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	// 解析命令行参数
	inputFile := flag.String("input", "", "Input CSV file")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Please specify the input CSV file using the -input option.")
		return
	}

	// 使用 bufio 读取整行输入
	readerInput := bufio.NewReader(os.Stdin)

	// 提示用户输入
	fmt.Print("Enter the application label to filter (press Enter to skip): ")
	appLabel, _ := readerInput.ReadString('\n')
	appLabel = strings.TrimSpace(appLabel) // 去除多余空格

	fmt.Print("The label needs to be replaced: ")
	oldLabel, _ := readerInput.ReadString('\n')
	oldLabel = strings.TrimSpace(oldLabel) // 去除多余空格

	fmt.Print("The new label: ")
	newLabel, _ := readerInput.ReadString('\n')
	newLabel = strings.TrimSpace(newLabel) // 去除多余空格

	// 打开CSV文件
	file, err := os.Open(*inputFile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV file:", err)
		return
	}

	if len(records) < 2 {
		fmt.Println("CSV file seems to be empty or only contains header.")
		return
	}

	// 找出scope列和href列
	scopeIndex := -1
	rulesetIndex := -1
	hrefIndex := -1
	for i, header := range records[0] {
		switch strings.ToLower(header) {
		case "scope":
			scopeIndex = i
		case "ruleset_name":
			rulesetIndex = i
		case "href":
			hrefIndex = i
		}
	}

	if scopeIndex == -1 {
		fmt.Println("Error: 'scope' column not found in the CSV file.")
		return
	}

	if rulesetIndex == -1 {
		fmt.Println("Error: 'ruleset_name' column not found in the CSV file.")
		return
	}

	// 创建新的记录集并去除href列
	changedRecords := [][]string{}
	header := removeColumn(records[0], hrefIndex)
	changedRecords = append(changedRecords, header)

	// 替换scope列中的label并生成新的记录，只保留匹配的行
	labelFound := false
	appLabelFound := false
	for _, record := range records[1:] {
		scope := record[scopeIndex]
		kvPairs := strings.Split(scope, ";")
		modified := false
		appMatch := appLabel == "" // 如果未指定appLabel，默认匹配所有
		for i, kv := range kvPairs {
			parts := strings.SplitN(kv, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				if key == "app" {
					if appLabel == "" || value == appLabel {
						appMatch = true
						appLabelFound = true
					}
				}

				if value == oldLabel {
					kvPairs[i] = key + ":" + newLabel
					labelFound = true
					modified = true
				}
			}
		}
		if modified && appMatch {
			record[scopeIndex] = strings.Join(kvPairs, ";")

			// 检查ruleset_name是否包含oldLabel
			if strings.Contains(record[rulesetIndex], oldLabel) {
				record[rulesetIndex] = strings.Replace(record[rulesetIndex], oldLabel, newLabel, 1)
			} else {
				record[rulesetIndex] += "-duplicate"
			}

			// 去掉href列并添加到结果集
			record = removeColumn(record, hrefIndex)
			changedRecords = append(changedRecords, record)
		}
	}

	if !appLabelFound && appLabel != "" {
		fmt.Println("Error: The specified application label was not found in the 'scope' column.")
		return
	}

	if !labelFound {
		fmt.Println("Error: The specified label was not found in the 'scope' column.")
		return
	}

	// 创建新文件名
	outputFile := strings.TrimSuffix(*inputFile, ".csv") + "_new.csv"
	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Error creating new CSV file:", err)
		return
	}
	defer outFile.Close()

	// 写入新的CSV文件
	writer := csv.NewWriter(outFile)
	err = writer.WriteAll(changedRecords)
	if err != nil {
		fmt.Println("Error writing to new CSV file:", err)
		return
	}

	fmt.Println("New CSV file created successfully:", outputFile)
}

// removeColumn 移除记录中的指定列
func removeColumn(record []string, index int) []string {
	if index < 0 || index >= len(record) {
		return record
	}
	return append(record[:index], record[index+1:]...)
}
