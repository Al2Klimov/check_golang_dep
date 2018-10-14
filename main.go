//go:generate go run vendor/github.com/Al2Klimov/go-gen-source-repos/main.go github.com/Al2Klimov/check_golang_dep

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/Al2Klimov/go-exec-utils"
	_ "github.com/Al2Klimov/go-gen-source-repos"
	. "github.com/Al2Klimov/go-monplug-utils"
	"github.com/golang/dep"
	. "github.com/otiai10/copy"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

func main() {
	os.Exit(ExecuteCheck(onTerminal, checkGolangDep))
}

func onTerminal() (output string) {
	return fmt.Sprintf(
		"For the terms of use, the source code and the authors\n"+
			"see the projects this program is assembled from:\n\n  %s\n",
		strings.Join(GithubcomAl2klimovGo_gen_source_repos, "\n  "),
	)
}

func checkGolangDep() (output string, perfdata PerfdataCollection, errs map[string]error) {
	if len(os.Args) != 3 {
		return "", nil, map[string]error{"Usage": errors.New(os.Args[0] + " GO_PACKAGE CACHE_DIR")}
	}

	signal.Ignore(syscall.SIGTERM)

	var project1 loadProjectAsyncOut
	var project2 updateAndLoadCopyOut

	{
		chCP := make(chan cacheProjectOut, 1)
		chPS := make(chan prepareSandboxOut, 1)
		var chLP chan loadProjectAsyncOut = nil
		var chUALC chan updateAndLoadCopyOut = nil

		go cacheProject(chCP)
		go prepareSandbox(chPS)

		var cachedProject cacheProjectOut
		var preparedSandbox prepareSandboxOut

		for {
			select {
			case cachedProject = <-chCP:
				if cachedProject.errs == nil {
					if errs == nil {
						chLP = make(chan loadProjectAsyncOut, 1)
						go loadProjectAsync(cachedProject.pkgDir, cachedProject.goPath, chLP)

						if chPS == nil {
							chUALC = make(chan updateAndLoadCopyOut, 1)
							go updateAndLoadCopy(cachedProject.pkgDir, cachedProject.goPath, preparedSandbox, chUALC)
						}
					}
				} else if errs == nil {
					errs = cachedProject.errs
				} else {
					for context, err := range cachedProject.errs {
						errs[context] = err
					}
				}

				chCP = nil
			case preparedSandbox = <-chPS:
				if preparedSandbox.errs == nil {
					defer os.RemoveAll(preparedSandbox.tmpDir)

					if errs == nil && chCP == nil {
						chUALC = make(chan updateAndLoadCopyOut, 1)
						go updateAndLoadCopy(cachedProject.pkgDir, cachedProject.goPath, preparedSandbox, chUALC)
					}
				} else if errs == nil {
					errs = preparedSandbox.errs
				} else {
					for context, err := range preparedSandbox.errs {
						errs[context] = err
					}
				}

				chPS = nil
			case project1 = <-chLP:
				if project1.err != nil {
					if errs == nil {
						errs = map[string]error{cachedProject.pkgDir: project1.err}
					} else {
						errs[cachedProject.pkgDir] = project1.err
					}
				}

				chLP = nil
			case project2 = <-chUALC:
				if project2.errs != nil {
					if errs == nil {
						errs = project2.errs
					} else {
						for context, err := range project2.errs {
							errs[context] = err
						}
					}
				}

				chUALC = nil
			}

			if chCP == nil && chPS == nil && chLP == nil && chUALC == nil {
				break
			}
		}
	}

	if errs == nil {
		output, perfdata = diffProjects(project1.project, project2.project)

		if output == "" {
			output = "Everything is up-to-date"
		} else {
			output = "Some dependencies aren't up-to-date!\n\n" + output + "\n"
		}
	}

	return
}

type cacheProjectOut struct {
	goPath, pkgDir string
	errs           map[string]error
}

func cacheProject(ch chan cacheProjectOut) {
	cacheDir := os.Args[2]

	if errMA := os.MkdirAll(cacheDir, 0700); errMA != nil {
		ch <- cacheProjectOut{errs: map[string]error{FormatCmd("mkdir", []string{"-p", cacheDir}, nil): errMA}}
		return
	}

	var goPath string

	{
		realPath, errES := filepath.EvalSymlinks(cacheDir)
		if errES != nil {
			ch <- cacheProjectOut{errs: map[string]error{FormatCmd("readlink", []string{"-m", cacheDir}, nil): errES}}
			return
		}

		goPath = path.Join(realPath, "go")
	}

	goPkg := os.Args[1]
	goCmdEnv := map[string]string{"LC_ALL": "C", "PATH": os.Getenv("PATH"), "GOPATH": goPath}

	if cmd, _, err := System("go", []string{"get", "-insecure", "-u", goPkg}, goCmdEnv, "/"); err != nil {
		ch <- cacheProjectOut{errs: map[string]error{cmd: err}}
		return
	}

	pkgDir := ""

	if cmd, out, err := System("go", []string{"list", "-json", goPkg}, goCmdEnv, "/"); err == nil {
		var unJson interface{}
		if json.Unmarshal(out, &unJson) == nil {
			if rootObject, rootIsObject := unJson.(map[string]interface{}); rootIsObject {
				if dirString, dirIsString := rootObject["Dir"].(string); dirIsString {
					pkgDir = dirString
				}
			}
		}

		if pkgDir == "" {
			ch <- cacheProjectOut{errs: map[string]error{cmd: errors.New("bad output")}}
			return
		}
	} else {
		ch <- cacheProjectOut{errs: map[string]error{cmd: err}}
		return
	}

	if cmd, _, err := System("dep", []string{"ensure", "-vendor-only"}, goCmdEnv, pkgDir); err != nil {
		ch <- cacheProjectOut{errs: map[string]error{cmd: err}}
		return
	}

	ch <- cacheProjectOut{goPath: goPath, pkgDir: pkgDir}
}

type prepareSandboxOut struct {
	tmpDir, goPath string
	errs           map[string]error
}

func prepareSandbox(ch chan prepareSandboxOut) {
	tmpDir, errTD := ioutil.TempDir("", "")
	if errTD != nil {
		ch <- prepareSandboxOut{errs: map[string]error{FormatCmd("mktemp", []string{"-d"}, nil): errTD}}
		return
	}

	realPath, errES := filepath.EvalSymlinks(tmpDir)
	if errES != nil {
		ch <- prepareSandboxOut{errs: map[string]error{FormatCmd("readlink", []string{"-m", tmpDir}, nil): errES}}
		os.RemoveAll(tmpDir)
		return
	}

	ch <- prepareSandboxOut{tmpDir: tmpDir, goPath: path.Join(realPath, "go")}
}

type updateAndLoadCopyOut struct {
	project *dep.Project
	errs    map[string]error
}

func updateAndLoadCopy(pkgDir, goPath string, sandbox prepareSandboxOut, ch chan updateAndLoadCopyOut) {
	if errCp := Copy(goPath, sandbox.goPath); errCp != nil {
		ch <- updateAndLoadCopyOut{errs: map[string]error{FormatCmd("cp", []string{"-r", goPath, sandbox.goPath}, nil): errCp}}
		return
	}

	pkgDir2 := path.Join(sandbox.goPath, strings.TrimPrefix(pkgDir, goPath))

	if cmd, _, err := System("dep", []string{"ensure", "-update"}, map[string]string{"LC_ALL": "C", "PATH": os.Getenv("PATH"), "GOPATH": sandbox.goPath}, pkgDir2); err != nil {
		ch <- updateAndLoadCopyOut{errs: map[string]error{cmd: err}}
		return
	}

	project, errLP := loadProject(pkgDir2, sandbox.goPath)
	if errLP != nil {
		ch <- updateAndLoadCopyOut{errs: map[string]error{pkgDir2: errLP}}
		return
	}

	ch <- updateAndLoadCopyOut{project: project}
}
