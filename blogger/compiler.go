package blogger

import (
	"os"
	"vsartor.com/gochs/log"
)

func CompileSource(srcDir, dstDir string, prod bool) error {
	log.Info("Starting compilation.")

	if !isDir(srcDir) {
		return log.Err("srcDir not a directory: <b>%s<r>", srcDir)
	}

	err := prepDstDir(srcDir, dstDir)
	if err != nil {
		return err
	}

	err = buildPages(srcDir, dstDir, prod)
	if err != nil {
		return err
	}

	// TODO: if prod, zip the output

	log.Info("Compilation complete.")
	return nil
}

func prepDstDir(srcDir, dstDir string) error {
	// if directory already exists, remove it
	if isDir(dstDir) {
		log.Warn("Deleting all contents of <b>%s<r>", dstDir)
		err := os.RemoveAll(dstDir)
		if err != nil {
			return log.Err("failed to delete: %s", err.Error())
		}
	}

	// create directory
	err := os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		return log.Err("failed to create dstDir")
	}

	// copy contents of static folder to new directory
	err = cpDir(srcDir+"/static/.", dstDir)
	if err != nil {
		return log.Err("failed to copy srcDir/static into dstDir: %s", err.Error())
	}

	return nil
}
