package board

import (
	"fmt"

	"github.com/Grar00t/khz-cli/internal/checkrecord"
	"github.com/Grar00t/khz-cli/internal/gitstate"
	"github.com/Grar00t/khz-cli/internal/model"
	"github.com/Grar00t/khz-cli/internal/policy"
)

type Entry struct {
	State  model.State `json:"state"`
	Name   string      `json:"name"`
	Detail string      `json:"detail"`
}

type Board struct {
	Entries       []Entry        `json:"entries"`
	Decision      model.Decision `json:"decision"`
	LatestReceipt string         `json:"latest_receipt,omitempty"`
}

func Build(g gitstate.Snapshot, p policy.CheckResult, checks []checkrecord.Record, latest string) Board {
	entries := []Entry{}
	if !g.InsideRepo {
		entries = append(entries, Entry{State: model.Info, Name: "source", Detail: "not inside Git repository"})
	} else if g.Clean {
		entries = append(entries, Entry{State: model.OK, Name: "source", Detail: fmt.Sprintf("clean %s", branchLabel(g))})
	} else {
		entries = append(entries, Entry{State: model.Warn, Name: "source", Detail: fmt.Sprintf("dirty staged=%d modified=%d untracked=%d", g.Staged, g.Modified, g.Untracked)})
	}
	if p.Valid {
		entries = append(entries, Entry{State: model.OK, Name: "policy", Detail: "valid"})
	} else {
		entries = append(entries, Entry{State: p.State, Name: "policy", Detail: firstPolicyDetail(p)})
	}
	if len(checks) == 0 {
		entries = append(entries, Entry{State: model.Skip, Name: "checks", Detail: "not measured"})
	} else {
		for _, c := range checks {
			entries = append(entries, Entry{State: c.Status, Name: c.Name, Detail: fmt.Sprintf("exit=%d duration=%dms", c.ExitCode, c.DurationMS)})
		}
	}
	states := make([]model.State, len(entries))
	for i := range entries {
		states[i] = entries[i].State
	}
	return Board{Entries: entries, Decision: model.Decide(states), LatestReceipt: latest}
}

func branchLabel(g gitstate.Snapshot) string {
	if g.Detached {
		return "detached"
	}
	if g.Unborn {
		return "unborn:" + g.Branch
	}
	if g.Branch != "" {
		return g.Branch
	}
	return "repository"
}

func firstPolicyDetail(p policy.CheckResult) string {
	if len(p.Events) > 0 {
		return p.Events[0].Detail
	}
	return "policy unavailable"
}
