package main

import (
	"bytes"
	"fmt"
	. "github.com/Al2Klimov/go-monplug-utils"
	"github.com/golang/dep"
	"github.com/golang/dep/gps"
	"github.com/golang/dep/gps/verify"
	"log"
	"math"
	"sort"
)

var posInf = math.Inf(1)

type nullWriter struct {
}

func (nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func loadProject(dir, goPath string) (*dep.Project, error) {
	return (&dep.Ctx{
		dir,
		goPath,
		[]string{goPath},
		dir,
		log.New(nullWriter{}, "", log.LstdFlags),
		log.New(nullWriter{}, "", log.LstdFlags),
		false,
		true,
		"",
		0,
	}).LoadProject()
}

type loadProjectAsyncOut struct {
	project *dep.Project
	err     error
}

func loadProjectAsync(dir, goPath string, ch chan loadProjectAsyncOut) {
	project, err := loadProject(dir, goPath)
	ch <- loadProjectAsyncOut{project, err}
}

type revision struct {
	version, digest string
}

func (r revision) String() string {
	return fmt.Sprintf("%s [ %s ]", r.version, r.digest)
}

func newRevision(lp gps.LockedProject) (rev revision) {
	rev.version = lp.Version().String()

	if vp, isVP := lp.(verify.VerifiableProject); isVP {
		rev.digest = vp.Digest.String()
	} else {
		rev.digest = "0:00000000000000000000000000000000"
	}

	return
}

type dependency struct {
	lhs, rhs revision
}

func diffProjects(lhs, rhs *dep.Project) (diff string, perfdata PerfdataCollection) {
	dependencies := map[string]*dependency{}

	for _, lp := range lhs.Lock.Projects() {
		dependencies[lp.Ident().String()] = &dependency{lhs: newRevision(lp)}
	}

	for _, lp := range rhs.Lock.Projects() {
		id := lp.Ident().String()

		if dep, hasDep := dependencies[id]; hasDep {
			dep.rhs = newRevision(lp)
		} else {
			dependencies[id] = &dependency{rhs: newRevision(lp)}
		}
	}

	orderedChangedDeps := []string{}
	unchangedDeps := 0

	for id, dep := range dependencies {
		if dep.rhs == dep.lhs {
			unchangedDeps++
		} else {
			orderedChangedDeps = append(orderedChangedDeps, id)
		}
	}

	sort.Strings(orderedChangedDeps)

	buf := &bytes.Buffer{}
	oldDeps := 0
	newDeps := 0
	updatedDeps := 0

	for _, id := range orderedChangedDeps {
		dep := dependencies[id]

		if dep.lhs == (revision{}) {
			fmt.Fprintf(buf, "+ %s @ %s\n", id, dep.rhs)
			newDeps++
		} else if dep.rhs == (revision{}) {
			fmt.Fprintf(buf, "- %s @ %s\n", id, dep.lhs)
			oldDeps++
		} else {
			fmt.Fprintf(buf, "- %s @ %s\n+ %s @ %s\n", id, dep.lhs, id, dep.rhs)
			updatedDeps++
		}
	}

	return buf.String(), PerfdataCollection{
		Perfdata{
			Label: "unchanged",
			Value: float64(unchangedDeps),
			Min:   OptionalNumber{true, 0.0},
		},
		Perfdata{
			Label: "updated",
			Value: float64(updatedDeps),
			Crit:  OptionalThreshold{true, true, 1.0, posInf},
			Min:   OptionalNumber{true, 0.0},
		},
		Perfdata{
			Label: "added",
			Value: float64(newDeps),
			Crit:  OptionalThreshold{true, true, 1.0, posInf},
			Min:   OptionalNumber{true, 0.0},
		},
		Perfdata{
			Label: "deleted",
			Value: float64(oldDeps),
			Warn:  OptionalThreshold{true, true, 1.0, posInf},
			Min:   OptionalNumber{true, 0.0},
		},
	}
}