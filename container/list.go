package container

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"go.uber.org/zap"
)

func ListContainer() {
	dirPath := fmt.Sprintf(DefaultInfoPath, "")
	dirPath = dirPath[:len(dirPath)-1]

	// ls dirPath
	files, err := os.ReadDir(dirPath)
	if err != nil {
		zap.L().Sugar().Errorf("read the directory error %v", err)
		return
	}
	containers := []*ContainerMeta{}
	for _, file := range files {
		meta, err := getContainerInfo(file)
		if meta == nil && err == nil {
			continue
		}
		if err != nil {
			zap.L().Sugar().Error("get container information error")
			continue
		}
		containers = append(containers, meta)
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tPID\t\\STATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
			item.ID,
			item.Name,
			item.PID,
			item.Status,
			item.Command,
			item.CreateAt.Format(time.DateTime),
		)
	}
	if err := w.Flush(); err != nil {
		zap.L().Sugar().Errorf("flush error %v", err)
		return
	}
}

func getContainerInfo(file fs.DirEntry) (*ContainerMeta, error) {
	if file.IsDir() {
		cfgPath := fmt.Sprintf(DefaultInfoPath, file.Name())
		f, err := os.ReadFile(filepath.Join(cfgPath, ConfigName))
		if err != nil {
			zap.L().Sugar().Errorf("read the container config file error %v", err)
			return nil, err
		}
		// decode
		containerMeta := new(ContainerMeta)
		err = json.Unmarshal(f, containerMeta)
		if err != nil {
			zap.L().Sugar().Error("unmarshal json error %v", err)
			return nil, err
		}
		return containerMeta, nil
	}
	return nil, nil
}
