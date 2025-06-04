package docker

type ExecParams struct {
	Args        []string
	Dettach     bool
	Env         []string
	Interactive bool
	Privileged  bool
	Tty         bool
	User        string
	Workdir     string
}
