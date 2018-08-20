package mgr

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/alibaba/pouch/pkg/utils"

	"github.com/sirupsen/logrus"
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
	// pouchEnvDir =
	pouchEnvDir = "/etc/profile.d"
	// this filepath is used in PouchContainer to store user input env via persistent file
	pouchEnvFile = "/etc/profile.d/pouchenv.sh"
)

// updateContainerEnv update the container's envs in
// /etc/profile.d/pouchenv.sh used by rich container.
func updateContainerEnv(inputRawEnv []string, baseFs string) error {
	var (
		envShPath = path.Join(baseFs, pouchEnvFile)
	)

	// check the existence of related files.
	// if dir of pouch.sh is not exist, it's unnecessary to update that files.
	if _, err := os.Stat(path.Join(baseFs, pouchEnvDir)); err != nil {
		return nil
	}
	if _, err := os.Stat(envShPath); err != nil {
		logrus.Warnf("failed to stat container's env file /etc/profile.d/pouchenv.sh: %v", err)
		return nil
	}

	inputEnv := utils.ConvertKVStrToMapWithNoErr(inputRawEnv)

	localEnv, err := getLocalEnv(envShPath)
	if err != nil {
		return fmt.Errorf("failed to get local env from file %s: %v", envShPath, err)
	}

	newEnv := combineLocalAndInputEnv(inputEnv, localEnv)

	// start to construct new content for file pouchenv.sh after combining
	// original local envs extracted from pouchenv.sh and input envs.
	var str string
	for key, value := range newEnv {
		str += fmt.Sprintf("export %s=\"%s\"\n", key, value)
	}
	ioutil.WriteFile(envShPath, []byte(str), 0755)

	return nil
}

// getLocalEnv gets local ENV from specified env shell files.
// Currently this is file /etc/profile.d/pouchenv.sh in the container.
// And every row of this file has a format of `export AAA="BBB"`.
func getLocalEnv(filename string) (map[string]string, error) {
	envMap := make(map[string]string)

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read env file %s: %v", filename, err)
	}

	rawData := string(b)
	envsh := strings.Trim(rawData, "\n")
	envs := strings.Split(envsh, "\n")
	for _, env := range envs {
		env = strings.TrimPrefix(env, "export ")
		key, value, err := utils.ConvertStrToKV(env)
		if err != nil {
			// TODO: just ignore the error or return err?
			return envMap, err
		}
		// FIXME: do we need to string.Trim(value, "\"")
		envMap[key] = value
	}

	return envMap, nil
}

// combineLocalAndInputEnv combines local envs extracted from pouchenv.sh and input envs.
// traverse env(key, value) from local file, and replace it if key also exists in input env.
// traverse env(key, value) from input env, if it does not exist in local env, just append it.
// FIXME: why not just only traverse input env?
func combineLocalAndInputEnv(inputEnv, localEnv map[string]string) map[string]string {
	for key, valueInLocal := range localEnv {
		valueInInput, exist := inputEnv[key]
		if !exist {
			// if key in the local env does not exist in input env,
			// it means this key will never be updated by input env.
			continue
		}
		// key in local env exists in input env.
		// TODO: explain this part by ziren
		s := strings.Replace(valueInInput, "\"", "\\\"", -1)
		valueInInput = strings.Replace(s, "$", "\\$", -1)

		if key == "PATH" {
			// FIXME: how to do if there is already a substr of `:$PATH` in itself.
			valueInInput = valueInInput + ":$PATH"
		}
		// two values are not equal, use the input one.
		if valueInInput != valueInLocal {
			logrus.Infof("env exists and the value is not same, key=%s, old value=%s, new value=%s", key, valueInLocal, valueInInput)
			localEnv[key] = valueInInput
		}
	}
	// append the brand new input env to local env map.
	for key, value := range inputEnv {
		if _, ok := localEnv[key]; !ok {
			logrus.Infof("env does not exist, set new key value pair, new key=%s, new value=%s", key, value)
			localEnv[key] = value
		}
	}
	return localEnv
}
