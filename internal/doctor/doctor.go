package doctor

import (
	"os/exec"
	"runtime"
)

type Tool struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
	Path      string `json:"path,omitempty"`
}

type Report struct {
	KHZVersion string `json:"khz_version"`
	GoVersion  string `json:"go_version"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	Tools      []Tool `json:"tools"`
}

func Build(version string) Report {
	tools := []Tool{}
	for _, name := range []string{"git", "gh", "pwsh", "powershell"} {
		path, err := exec.LookPath(name)
		tools = append(tools, Tool{Name: name, Available: err == nil, Path: path})
	}
	return Report{KHZVersion: version, GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, Tools: tools}
}
