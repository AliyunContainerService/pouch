package mgr

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// container env is used to manage environment variables in container.
// In original ways, processes in container can only accepted env when startup.
// And container engine would store those data in container.json which is persistent on local disk.
// There is no env ralated data in container or image.
//
// While for rich container mode, env management in container is a little bit complicated.
// First, introduce some scenarios in practice:
// * init process in rich container will not pass envs to user-specific CMD process if user switched from root;
// * newly created shell command should herit all envs of the container, so all envs should be stored in dir
// /etc/profild.d/*.sh (here is pouchenv.sh). After that a process execed from outside would take env of init process,
// then such action like reload of detailed user application could share the env.
const (
	// this filepath is used in PouchContainer to store user input env via persistent file
	pouchEnvFile = "/etc/profile.d/pouchenv.sh"
)

// updateContainerEnv update the container's envs in
// /etc/profile.d/pouchenv.sh used by rich container.
func updateContainerEnv(env []string, baseFs string) error {
	// if target env directory is not exist, return
	// generally speaking, we should check env file exist to decide
	// whether to update env file, but in case of missing something,
	// we check directory existence here to make a decision.
	if _, err := os.Stat(filepath.Join(baseFs, filepath.Dir(pouchEnvFile))); err != nil {
		return nil
	}

	var (
		str string
	)
	for _, kv := range env {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 1 {
			parts = append(parts, "")
		}
		if len(parts[1]) > 0 && !strings.Contains(parts[0], ".") {
			s := strings.Replace(parts[1], "\\", "\\\\", -1)
			s = strings.Replace(s, "\"", "\\\"", -1)
			s = strings.Replace(s, "$", "\\$", -1)
			if parts[0] == "PATH" {
				s = parts[1] + ":$PATH"
			}
			str += fmt.Sprintf("export %s=\"%s\"\n", parts[0], s)
		}
	}
	ioutil.WriteFile(filepath.Join(baseFs, pouchEnvFile), []byte(str), 0755)

	return nil
}
