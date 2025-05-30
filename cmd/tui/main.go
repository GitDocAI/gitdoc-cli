package main

import (
	"bufio"
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
	config := config.NewConfig()

	client_path, err := os.Getwd()
	if err != nil {
		fmt.Printf("An error ocurred: %s", err.Error())
		return
	}

	initializer := initializer.NewInitializer(&config)
	if err := initializer.CheckEnvironment(); err != nil {
		fmt.Println(err)
		if err := initializer.SetEnvironment(); err != nil {
			fmt.Println(err)
			return
		}
	}

	wc_service := service_adapter.NewWebConfigAdapter()
	wc_repo := repository_adapter.NewWebConfigRepository(config.ProjectPath, client_path)
	wc_controller := webconfig.NewWebConfigController(wc_repo, wc_service)
	error := wc_controller.InitializeWebConfig(client_path + "/site.config.json")
	if error != nil {
		fmt.Println(error)
		return
	}

	fmt.Println("Start project building....")
	build_cmd := exec.Command("npm", "run", "build")
	build_cmd.Dir = config.ProjectPath // Sets the working directory

	if err := build_cmd.Run(); err != nil {
		fmt.Println("Error building the project", err)
		return
	}
	fmt.Println("Project builded start running....")

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
			fmt.Printf("Project running on : http://localhost:%s", match[1])
			return
		}
	}
}
