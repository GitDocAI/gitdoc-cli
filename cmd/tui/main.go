package main

import (
	"bufio"
	"flag" // Import the flag package
	"fmt"
	"os"
	"os/exec"
	"regexp"

	service_adapter "github.com/YHVCorp/DockitAi/cli/internals/application/webconfig"
	"github.com/YHVCorp/DockitAi/cli/internals/config"
	"github.com/YHVCorp/DockitAi/cli/internals/infraestructure/initializer"
	repository_adapter "github.com/YHVCorp/DockitAi/cli/internals/infraestructure/webconfig"
	"github.com/YHVCorp/DockitAi/cli/internals/interface/controllers/webconfig"
)

func main() {
	// Define the flags
	inputPath := flag.String("input", "", "Determines the source code directory (defaults to current path)")
	outputPath := flag.String("output", "", "Determines the directory where the results will be saved")
	replaceOnly := flag.Bool("replace", false, "Generates app data without running the application")
	helpFlag := flag.Bool("help", false, "Displays help message for available flags")

	// Parse the flags
	flag.Parse()

	if *helpFlag {
		fmt.Println("Usage of dokcitAi-cli:")
		flag.PrintDefaults()
		return
	}

	config := config.NewConfig()

	// Determine client_path based on --input flag
	client_path, err := os.Getwd()
	if err != nil {
		fmt.Printf("An error occurred: %s\n", err.Error())
		return
	}
	if *inputPath != "" {
		client_path = *inputPath
	}

	initializer := initializer.NewInitializer(&config)
	if err := initializer.CheckEnvironment(); err != nil {
		fmt.Println(err)
		if err := initializer.SetEnvironment(); err != nil {
			fmt.Println(err)
			return
		}
	}

	savepath := config.ProjectPath
	if *outputPath != "" {
		savepath = *outputPath
	}
	wc_service := service_adapter.NewWebConfigAdapter()
	// Pass outputPath to the repository if needed, or handle it later for saving results
	wc_repo := repository_adapter.NewWebConfigRepository(savepath, client_path)
	wc_controller := webconfig.NewWebConfigController(wc_repo, wc_service)

	error := wc_controller.InitializeWebConfig(client_path + "/site.config.json")
	if error != nil {
		fmt.Println(error)
		return
	}

	if *replaceOnly {
		fmt.Println("App data generated successfully. Application will not be run due to --replace flag.")
		return
	}

	fmt.Println("Start project building....")
	build_cmd := exec.Command("npm", "run", "build")
	build_cmd.Dir = config.ProjectPath // Sets the working directory

	if err := build_cmd.Run(); err != nil {
		fmt.Println("Error building the project", err)
		return
	}
	fmt.Println("Project built, starting run....")

	cmd := exec.Command("npm", "run", "start")
	cmd.Dir = config.ProjectPath // Sets the working directory
	check_running_port(cmd)
	cmd.Wait()
}

func check_running_port(cmd *exec.Cmd) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("Error running the project:", err)
		return
	}

	scanner := bufio.NewScanner(stdoutPipe)
	re := regexp.MustCompile(`http://localhost:(\d+)`)
	for scanner.Scan() {
		line := scanner.Text()
		if match := re.FindStringSubmatch(line); match != nil {
			fmt.Printf("Project running on: http://localhost:%s\n", match[1])
			return
		}
	}
}
