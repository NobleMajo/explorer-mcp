package container

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/NobleMajo/explorer-mcp/internal/fsutil"
	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

type containerOverviewResponse struct {
	IsDockerAvailable          bool               `json:"isDockerAvailable"`
	IsPodmanAvailable          bool               `json:"isPodmanAvailable"`
	AvailableContainerCLICount int                `json:"availableContainerCLICount"`
	AvailableContainerCLINames []string           `json:"availableContainerCLINames"`
	DetectedContainerFileCount int                `json:"detectedContainerFileCount"`
	DetectedContainerFilePaths []string           `json:"detectedContainerFilePaths"`
	RunningContainerCount      int                `json:"runningContainerCount"`
	RunningContainers          []runningContainer `json:"runningContainers"`
}

type runningContainer struct {
	ContainerName string `json:"containerName"`
	ImageName     string `json:"imageName"`
	StatusText    string `json:"statusText"`
	RuntimeName   string `json:"runtimeName"`
}

func buildContainerOverview(verbose bool) (containerOverviewResponse, error) {
	_ = verbose
	root, err := os.Getwd()
	if err != nil {
		return containerOverviewResponse{}, err
	}

	resp := containerOverviewResponse{
		DetectedContainerFilePaths: []string{},
		RunningContainers:          []runningContainer{},
		AvailableContainerCLINames: []string{},
	}

	availableCLIs := detectAvailableContainerCLIs()
	resp.AvailableContainerCLINames = availableCLIs
	resp.AvailableContainerCLICount = len(availableCLIs)
	resp.IsDockerAvailable = slices.Contains(availableCLIs, "docker")
	resp.IsPodmanAvailable = slices.Contains(availableCLIs, "podman")

	paths, err := detectContainerFilePaths(root)
	if err != nil {
		return containerOverviewResponse{}, err
	}
	resp.DetectedContainerFilePaths = paths
	resp.DetectedContainerFileCount = len(paths)

	containers := make([]runningContainer, 0)
	for _, runtimeName := range globals.KnownContainerRuntimeCLINames {
		if !slices.Contains(availableCLIs, runtimeName) {
			continue
		}
		containers = append(containers, listRunningContainers(runtimeName, root)...)
	}
	resp.RunningContainers = containers
	resp.RunningContainerCount = len(containers)

	return resp, nil
}

func detectAvailableContainerCLIs() []string {
	available := make([]string, 0, len(globals.KnownContainerCLINames))
	for _, name := range globals.KnownContainerCLINames {
		if _, err := exec.LookPath(name); err == nil {
			available = append(available, name)
		}
	}
	return available
}

func detectContainerFilePaths(root string) ([]string, error) {
	seen := make(map[string]bool)
	paths := make([]string, 0)

	addPath := func(relPath string) {
		if relPath == "" || seen[relPath] {
			return
		}
		seen[relPath] = true
		paths = append(paths, relPath)
	}

	for _, pattern := range globals.KnownContainerFileNames {
		if strings.Contains(pattern, "*") {
			matches, err := filepath.Glob(filepath.Join(root, pattern))
			if err != nil {
				return nil, err
			}
			for _, match := range matches {
				if fsutil.FileExists(match) {
					addPath(filepath.Base(match))
				}
			}
			continue
		}

		if fsutil.FileExists(filepath.Join(root, pattern)) {
			addPath(pattern)
		}
	}

	for _, dirName := range globals.KnownContainerDirectoryNames {
		if fsutil.DirExists(filepath.Join(root, dirName)) {
			addPath(dirName + "/")
		}
	}

	devcontainerConfig := filepath.Join(".devcontainer", "devcontainer.json")
	if fsutil.FileExists(filepath.Join(root, devcontainerConfig)) {
		addPath(devcontainerConfig)
	}

	sort.Strings(paths)
	return paths, nil
}

func listRunningContainers(runtimeName, dir string) []runningContainer {
	cmd := exec.Command(runtimeName, "ps", "--format", "{{.Names}}\t{{.Image}}\t{{.Status}}")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return []runningContainer{}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	containers := make([]runningContainer, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		container := runningContainer{RuntimeName: runtimeName}
		if len(fields) > 0 {
			container.ContainerName = fields[0]
		}
		if len(fields) > 1 {
			container.ImageName = fields[1]
		}
		if len(fields) > 2 {
			container.StatusText = fields[2]
		}
		containers = append(containers, container)
	}

	return containers
}
