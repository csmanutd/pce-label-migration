package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	// Parse CLI arguments
	csvFile := flag.String("csv", "", "CSV file to process")
	oldRuleset := flag.String("old", "", "Old ruleset name to match")
	newRuleset := flag.String("new", "", "New ruleset name")
	flag.Parse()

	if *csvFile == "" || *oldRuleset == "" || *newRuleset == "" {
		fmt.Print("Please provide the CSV file, old ruleset name, and new ruleset name. ")
		flag.Usage()
		return
	}

	// Open the CSV file
	file, err := os.Open(*csvFile)
	if err != nil {
		fmt.Printf("Failed to open file: %s\n", err)
		return
	}
	defer file.Close()

	// Read the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Failed to read CSV file: %s\n", err)
		return
	}

	if len(records) == 0 {
		fmt.Print("CSV file is empty. ")
		return
	}

	// Identify the necessary columns
	headers := records[0]
	rulesetNameIndex := -1
	srcLabelsIndex := -1
	dstLabelsIndex := -1
	rulesetScopeIndex := -1
	columnsToInclude := []int{}
	columnsForAdditionalRules := []int{}

	for i, header := range headers {
		if header == "ruleset_name" {
			rulesetNameIndex = i
		}
		if header == "src_labels" {
			srcLabelsIndex = i
		}
		if header == "dst_labels" {
			dstLabelsIndex = i
		}
		if header == "ruleset_scope" {
			rulesetScopeIndex = i
		}
		if header != "ruleset_href" && header != "rule_href" {
			columnsToInclude = append(columnsToInclude, i)
			columnsForAdditionalRules = append(columnsForAdditionalRules, i)
		}
	}

	if rulesetNameIndex == -1 {
		fmt.Print("ruleset_name column not found in CSV file. ")
		return
	}

	if srcLabelsIndex == -1 || dstLabelsIndex == -1 {
		fmt.Print("src_labels or dst_labels column not found in CSV file. ")
		return
	}

	// Create a new CSV file to save the results
	outputFile := strings.TrimSuffix(*csvFile, ".csv") + "_new.csv"
	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Failed to create output file: %s\n", err)
		return
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// Write the headers to the new file, including 'ruleset_scope' but excluding 'ruleset_href' and 'rule_href'
	writer.Write(headers)

	// Process each row for the first part: copy old ruleset to new ruleset
	for _, record := range records[1:] {
		if record[rulesetNameIndex] == *oldRuleset {
			newRow := make([]string, len(headers))
			for _, i := range columnsToInclude {
				if i == rulesetNameIndex {
					newRow[i] = *newRuleset
				} else if i != rulesetScopeIndex {
					newRow[i] = record[i]
				} else {
					newRow[i] = "" // Keep the column but clear its contents
				}
			}
			writer.Write(newRow)
		}
	}

	fmt.Printf("Processing completed. New file saved as: %s\n", outputFile)

	// Ask if the user wants to copy additional rules
	readerInput := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want to copy and paste additional rules? (Y/n): ")
	copyMore, _ := readerInput.ReadString('\n')
	copyMore = strings.TrimSpace(copyMore)
	if copyMore == "" || strings.ToUpper(copyMore) == "Y" {
		var appLabel, origLabel, newLabel string
		fmt.Print("Input the application label to filter: ")
		appLabel, _ = readerInput.ReadString('\n')
		appLabel = strings.TrimSpace(appLabel)
		fmt.Print("Input the original label you want to replace: ")
		origLabel, _ = readerInput.ReadString('\n')
		origLabel = strings.TrimSpace(origLabel)
		fmt.Print("Input the new label: ")
		newLabel, _ = readerInput.ReadString('\n')
		newLabel = strings.TrimSpace(newLabel)

		// Prepare a regular expression for exact match replacement
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(origLabel) + `\b`)

		// Copy and modify rows based on the additional rules
		for _, record := range records[1:] {
			rulesetName := record[rulesetNameIndex]
			if rulesetName != *oldRuleset && rulesetName != *newRuleset {
				srcLabels := record[srcLabelsIndex]
				dstLabels := record[dstLabelsIndex]

				if strings.Contains(srcLabels, appLabel) && re.MatchString(srcLabels) ||
					(strings.Contains(dstLabels, appLabel) && re.MatchString(dstLabels)) {

					newRow := make([]string, len(headers))
					for _, i := range columnsForAdditionalRules {
						if i == srcLabelsIndex {
							newRow[i] = re.ReplaceAllString(record[i], newLabel)
						} else if i == dstLabelsIndex {
							newRow[i] = re.ReplaceAllString(record[i], newLabel)
						} else {
							newRow[i] = record[i]
						}
					}
					writer.Write(newRow)
				}
			}
		}
		fmt.Println("Additional rules copied and modified.")
	} else {
		fmt.Println("No additional rules copied.")
	}
}

