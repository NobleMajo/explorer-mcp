package container

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"sort"
	"strings"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

type containerOverviewResponse struct {
	CLIFound       []string            `json:"cliFound,omitempty"`
	ContainerFound map[string][]string `json:"containerFound,omitempty"`
}

func buildContainerOverview(verbose bool) (containerOverviewResponse, error) {
	_ = verbose
	root, err := os.Getwd()
	if err != nil {
		return containerOverviewResponse{}, err
	}

	cliFound := detectAvailableContainerCLIs()
	containerFound := make(map[string][]string)

	for _, runtimeName := range globals.KnownContainerRuntimeCLINames {
		if !slices.Contains(cliFound, runtimeName) {
			continue
		}
		entries := listRunningContainerEntries(runtimeName, root)
		if len(entries) > 0 {
			containerFound[runtimeName] = entries
		}
	}

	sort.Strings(cliFound)

	return containerOverviewResponse{
		CLIFound:       cliFound,
		ContainerFound: containerFound,
	}, nil
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

func listRunningContainerEntries(runtimeName, dir string) []string {
	cmd := exec.Command(runtimeName, "ps", "--format", "{{.Names}}\t{{.Image}}\t{{.Ports}}")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	entries := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		name := fieldAt(fields, 0)
		image := fieldAt(fields, 1)
		ports := fieldAt(fields, 2)
		mounts := containerMounts(runtimeName, dir, name)

		entries = append(entries, formatContainerEntry(name, image, ports, mounts))
	}

	sort.Strings(entries)
	return entries
}

func fieldAt(fields []string, index int) string {
	if index >= len(fields) {
		return ""
	}
	return strings.TrimSpace(fields[index])
}

func containerMounts(runtimeName, dir, containerName string) string {
	if containerName == "" {
		return ""
	}

	cmd := exec.Command(
		runtimeName,
		"inspect",
		"-f",
		"{{range .Mounts}}{{.Source}}:{{.Destination}};{{end}}",
		containerName,
	)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	mounts := strings.TrimSpace(string(out))
	return strings.TrimSuffix(mounts, ";")
}

func formatContainerEntry(name, image, ports, mounts string) string {
	return fmt.Sprintf("name@%s image@%s ports@%s mounts@%s", name, image, ports, mounts)
}
