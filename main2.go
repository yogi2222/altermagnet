package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// checkIfAmcheckInstalled checks if `amcheck` is installed
func checkIfAmcheckInstalled() bool {
	cmd := exec.Command("amcheck", "--help") // Check if the command runs successfully
	err := cmd.Run()
	return err == nil
}

// installAmcheck installs `amcheck` using `pip`
func installAmcheck() error {
	fmt.Println("Installing amcheck...")
	cmd := exec.Command("pip3", "install", "amcheck")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Special elements that follow "u d u d" pattern
var specialElements = map[string]bool{
	"Ti": true, "V": true, "Cr": true, "Mn": true, "Fe": true,
	"Co": true, "Ni": true, "Cu": true, "Mo": true, "Ru": true,
}

// / generateCombinations generates all valid sequences with n/2 'u' and n/2 'd'.
func generateCombinations(n int) []string {
	if n%2 != 0 {
		return nil // n must be even
	}

	var result []string
	seen := make(map[string]bool)
	var backtrack func(path []rune, uCount, dCount int)

	backtrack = func(path []rune, uCount, dCount int) {
		if len(path) == n {
			seq := string(path)
			flipped := flip(seq)
			if !seen[seq] && !seen[flipped] {
				seen[seq] = true
				result = append(result, formatOutput(seq))
			}
			return
		}
		if uCount < n/2 {
			backtrack(append(path, 'u'), uCount+1, dCount)
		}
		if dCount < n/2 {
			backtrack(append(path, 'd'), uCount, dCount+1)
		}
	}

	backtrack([]rune{}, 0, 0)
	sort.Strings(result) // Ensure sorted output
	return result
}

// flip swaps 'u' with 'd' and vice versa.
func flip(s string) string {
	runes := []rune(s)
	for i, r := range runes {
		if r == 'u' {
			runes[i] = 'd'
		} else {
			runes[i] = 'u'
		}
	}
	return string(runes)
}

// formatOutput formats the output to be space-separated.
func formatOutput(s string) string {
	var result string
	for i, r := range s {
		if i > 0 {
			result += " "
		}
		result += string(r)
	}
	return result
}

func runNnAmCheck(vaspFilePath string, logFile *os.File) ([]string, []int) {
	cmd := exec.Command("amcheck", vaspFilePath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("error creating stdin pipe:", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("error creating stdout pipe", err)
	}
	if err := cmd.Start(); err != nil {
		fmt.Println("error starting command", err)
	}
	scanner := bufio.NewScanner(stdout)
	// Regex to extract element type and atom positions
	elementRegex := regexp.MustCompile(`Orbit of (\w+) atoms at positions:`)
	atomRegex := regexp.MustCompile(`\d+ \(\d+\) \[\s*[-?\d.eE]+\s+[-?\d.eE]+\s+[-?\d.eE]+\s*\]`)
	var element string
	var atomCount int
	var elementArray []string
	var atomCountArray []int
	go func() {
		defer stdin.Close()
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line) // Print amcheck output to console
			logFile.WriteString(fmt.Sprintf("%s\n", line))

			// Detect the primitive unit cell prompt and respond with "Y"
			if strings.Contains(line, "Do you want to use it instead? (Y/n)") {
				fmt.Fprintln(stdin, "Y")
			}

			// Check if this line mentions the element being processed
			if matches := elementRegex.FindStringSubmatch(line); matches != nil {
				element = matches[1] // Extract element name (e.g., "Al")
				atomCount = 0        // Reset atom count
			}

			// Count atom lines
			if atomRegex.MatchString(line) {
				atomCount++
			}
			// Detect when amcheck requests input
			if strings.Contains(line, "Type spin (u, U, d, D, n, N, nn or NN) for each of them") {
				elementArray = append(elementArray, element)
				atomCountArray = append(atomCountArray, atomCount)
				var response []string

				if specialElements[element] {
					// Alternate between "u d u d" for special elements
					options := []string{"u", "d"}
					for i := 0; i < atomCount; i++ {
						response = append(response, options[i%2])
					}
				} else {
					response = append(response, "nn")
				}

				// response = append(response, "nn")

				// Send the computed response
				responseStr := strings.Join(response, " ") + "\n"
				fmt.Println("Sending response:", responseStr)
				logFile.WriteString(fmt.Sprintf("Sending response: %s\n", responseStr))
				stdin.Write([]byte(responseStr))
			}
		}
	}()
	if err := cmd.Wait(); err != nil {
		fmt.Println("amcheck command finished with error:", err)
	}
	return elementArray, atomCountArray
}

func runAmCheck(vaspFilePath string, inputArray []string, logFile *os.File) string {
	cmd := exec.Command("amcheck", vaspFilePath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("error creating stdin pipe:", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("error creating stdout pipe", err)
	}
	if err := cmd.Start(); err != nil {
		fmt.Println("error starting command", err)
	}
	index := 0
	scanner := bufio.NewScanner(stdout)

	// Regex to extract element type and atom positions
	// elementRegex := regexp.MustCompile(`Orbit of (\w+) atoms at positions:`)
	//atomRegex := regexp.MustCompile(`\d+ \(\d+\) \[[-?\d.]+ [-?\d.]+ [-?\d.]+\]`)
	altermagnetRegex := regexp.MustCompile(`Altermagnet\?\s*(True|False)`) // Detect "Altermagnet? True/False"

	// var element string
	//var atomCount int
	var altermagnetResult string

	go func() {
		defer stdin.Close()
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line) // Print amcheck output to console
			logFile.WriteString(fmt.Sprintf("%s\n", line))
			// Check if this line mentions the element being processed
			// if matches := elementRegex.FindStringSubmatch(line); matches != nil {
			// 	element = matches[1] // Extract element name (e.g., "Al")
			// 	atomCount = 0        // Reset atom count
			// }

			// Count atom lines
			//if atomRegex.MatchString(line) {
			//	atomCount++
			//}

			// Detect the primitive unit cell prompt and respond with "Y"
			if strings.Contains(line, "Do you want to use it instead? (Y/n)") {
				fmt.Fprintln(stdin, "Y")
			}

			// Detect Altermagnet result
			if matches := altermagnetRegex.FindStringSubmatch(line); matches != nil {
				altermagnetResult = matches[1] // Extract "True" or "False"
			}

			// Detect when amcheck requests input
			if strings.Contains(line, "Type spin (u, U, d, D, n, N, nn or NN) for each of them") {
				// var response []string
				// if specialElements[element] {
				// 	// Alternate between "u d u d" for special elements
				// 	options := []string{"u", "d"}
				// 	for i := 0; i < atomCount; i++ {
				// 		response = append(response, options[i%2])
				// 	}
				// } else {
				// 	response = append(response, "nn")
				// }

				// Send the computed response
				responseStr := inputArray[index] + "\n"
				index++
				fmt.Println("Sending response:", responseStr)
				logFile.WriteString(fmt.Sprintf("Sending response: %s\n", responseStr))
				stdin.Write([]byte(responseStr))
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		fmt.Println("amcheck command finished with error:", err)
	}

	// Print the final result of Altermagnet
	fmt.Println("Altermagnet Result:", altermagnetResult)
	logFile.WriteString(fmt.Sprintf("Altermagnet Result: %s\n", altermagnetResult))
	return altermagnetResult

}

// Function to generate all combinations
func generateInputCombinations(vaspFilePath string, arr [][]string, index int, current []string, logFile *os.File) (int, int) {
	total := 0
	trueCount := 0

	var helper func(int, []string)
	helper = func(idx int, curr []string) {
		if idx == len(arr) {
			total++
			output := runAmCheck(vaspFilePath, curr, logFile)
			if output == "True" {
				trueCount++
			}
			return
		}
		for _, val := range arr[idx] {
			helper(idx+1, append(curr, val))
		}
	}

	helper(index, current)
	return total, trueCount
}


func main() {
	//cmd := exec.Command("bash", "-c", "source /Users/aryang/Documents/python_repos/Yogi/py-amcheck/bin/activate && echo 'Virtual environment activated'")
	//
	//// Use current shell environment
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	//
	//// Run the command
	//err := cmd.Run()
	//if err != nil {
	//	fmt.Println("Error:", err)
	//}
	//time.Sleep(10 * time.Second)
	if !checkIfAmcheckInstalled() {
		err_ := installAmcheck()
		if err_ != nil {
			fmt.Println("Failed to install amcheck:", err_)
			return
		}
		fmt.Println("amcheck installed successfully!")
	} else {
		fmt.Println("amcheck is already installed.")
	}

	// Generate unique `u/d` sequences for atom counts up to 12
	maxN := 26
	sequenceMap := make(map[int][]string)
	for n := 2; n <= maxN; n += 2 {
		sequenceMap[n] = generateCombinations(n)
		// for _, r := range sequenceMap[n] {
		// 	fmt.Println(r)
		// }
	}

	// Ask user for the folder path
	fmt.Print("Enter folder path: ")
	var folderPath string
	fmt.Scanln(&folderPath)

	// Read directory contents
	files, err := os.ReadDir(folderPath)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}
	baseDir := filepath.Dir(folderPath)
	// Define "trueFiles" folder path
	trueFilesDir := filepath.Join(baseDir, "trueFiles")

	// Ensure "trueFiles" folder exists
	if err := os.MkdirAll(trueFilesDir, os.ModePerm); err != nil {
		fmt.Println("Error creating trueFiles folder:", err)
		return
	}
	logFilePath := filepath.Join(baseDir, "amcheck_log.txt")
	err = os.Remove(logFilePath)
	if err != nil {
		fmt.Println("Error deleting log file:", err)
	}
	logFile, err := os.Create(logFilePath)
	if err != nil {
		fmt.Println("Error creating log file:", err)
		return
	}
	defer logFile.Close()
	// Print file names with full paths
	fmt.Println("Files in folder:")
	for index, file := range files {
		fullPath := filepath.Join(folderPath, file.Name())
		runAmCheckOnFile(sequenceMap, fullPath, trueFilesDir, logFile)
		logFile.WriteString(fmt.Sprintf("File Number Processed: %d\n", index))
		fmt.Println("File Number Processed: ", index)
	}
}

func runAmCheckOnFile(sequenceMap map[int][]string, vaspFilePath string, trueFilesDir string, logFile *os.File) {
	//run NN amcheck
	elementArray, atomCountArray := runNnAmCheck(vaspFilePath, logFile)
	logFile.WriteString(fmt.Sprintf("Count of element is %v %v /n", elementArray, atomCountArray))
	fmt.Println("Count of element is", elementArray, atomCountArray)

	var elementInputArray [][]string
	for index, element := range elementArray {
		if specialElements[element] {
			if atomCountArray[index] > 24 {
				fmt.Println("Atom Count is Greater than 24. Try more combinations.")
				logFile.WriteString(fmt.Sprintf("Atom Count is Greater than 24. Try more combinations."))
				return
			}
			elementInputArray = append(elementInputArray, sequenceMap[atomCountArray[index]])
		} else {
			elementInputArray = append(elementInputArray, []string{"nn"})
		}
	}

	// Define destination path for the copied file
	destPath := filepath.Join(trueFilesDir, filepath.Base(vaspFilePath))

	totalComb, trueComb := generateInputCombinations(vaspFilePath, elementInputArray, 0, []string{}, logFile)

fmt.Printf("Out of %d combinations, %d returned Altermagnet? True\n", totalComb, trueComb)
logFile.WriteString(fmt.Sprintf("Out of %d combinations, %d returned Altermagnet? True\n", totalComb, trueComb))

if trueComb > 0 {
    fmt.Println("Altermagnet found in one or more configurations.")
    if err := copyFile(vaspFilePath, destPath); err != nil {
        fmt.Println("Error copying file:", err)
        return
    }
    fmt.Printf("File %s copied successfully to: %s\n", vaspFilePath, destPath)
} else {
    fmt.Println("No altermagnetism detected.")
}

}

// Function to copy file contents
func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy file contents
	_, err = io.Copy(dstFile, srcFile)
	return err
}
