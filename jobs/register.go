package jobs

import (
	"sort"
)

func (r *Runner) registerJobs() {
	health(r, r.log)

	var names []string
	for k := range r.jobs {
		names = append(names, k)
	}
	sort.Strings(names)

	r.log.Println("Registered jobs:", names)
}
