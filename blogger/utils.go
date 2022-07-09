package blogger

import (
	"os"
	"os/exec"
	"vsartor.com/gochs/log"
)

func isDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		// can't stat for some reason, probably "IsNotExist" but in practice irrelevant
		return false
	}

	return fi.IsDir()
}

func cpDir(src, dst string) error {
	log.Dbg("Running: <gr>cp -a %s %s<r>", src, dst)
	cmd := exec.Command("cp", "-a", src, dst)
	return cmd.Run()
}
