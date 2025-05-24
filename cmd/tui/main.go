package main

import (
	"fmt"
	"os"
	"os/exec"

	service_adapter "github.com/YHVCorp/DockitAi/cli/internals/application/webconfig"
	"github.com/YHVCorp/DockitAi/cli/internals/config"
	repository_adapter "github.com/YHVCorp/DockitAi/cli/internals/infraestructure/webconfig"
	"github.com/YHVCorp/DockitAi/cli/internals/interface/controllers/webconfig"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Must Provide CLient Path")
		return
	}

	config := config.NewConfig()
	client_path := os.Args[1]
	wc_service := service_adapter.NewWebConfigAdapter()
	wc_repo := repository_adapter.NewWebConfigRepository(config.ProjectPath, client_path)
	wc_controller := webconfig.NewWebConfigController(wc_repo, wc_service)
	error := wc_controller.InitializeWebConfig(client_path + "/client.json")
	if error != nil {
		fmt.Println(error)
		return
	}
	cmd := exec.Command("npm", "run", "dev")
	cmd.Dir = config.ProjectPath // Establece el working directory
	cmd.Stdout = os.Stdout       // Muestra la salida en consola
	cmd.Stderr = os.Stderr       // Muestra errores en consola

	if err := cmd.Run(); err != nil {
		fmt.Println("Error al correr el proyecto:", err)
	}
}
