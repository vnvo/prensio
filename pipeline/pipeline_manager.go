package pipeline

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"

	"github.com/siddontang/go-log/log"
	"github.com/vnvo/prensio/config"
)

type PipelineManager struct {
	pipelines []*CDCPipeline
	cdcConfs  []config.CDCConfig

	wg  *sync.WaitGroup
	ctx context.Context
}

func NewPipelineManager(confPath string) (*PipelineManager, error) {
	// iterate over confPath, find any toml file and create a pipeline for each one
	log.Infof("loading cfg files from '%s'", confPath)
	cdcConfs, err := config.NewCDCConfigList(confPath)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	wg := sync.WaitGroup{}

	pm := PipelineManager{
		make([]*CDCPipeline, 0),
		cdcConfs,
		&wg,
		ctx,
	}

	return &pm, nil
}

func (pm *PipelineManager) Start() {
	var err error
	for _, conf := range pm.cdcConfs {
		err = pm.startNewPipeline(conf)
		if err != nil {
			panic(err)
		}
	}
}

func (pm *PipelineManager) startNewPipeline(conf config.CDCConfig) error {

	newPipeline := NewCDCPipeline(conf.Mysql.Name, &conf, pm.wg)
	//log.Infof("[%s] pipeline created.", conf.Mysql.Name)
	log.Infof(
		"creating new pipeline for source=%s, sink=%s",
		conf.Mysql.Name,
		conf.KafkaSink.Name)

	err := newPipeline.Init()
	if err != nil {
		return err
	}

	pm.wg.Add(1)
	go func() {
		defer pm.wg.Done()
		newPipeline.Run(pm.ctx)
	}()

	pm.pipelines = append(pm.pipelines, &newPipeline)
	return nil
}

func getCfgFiles(confPath string) ([]string, error) {
	cfgFiles := make([]string, 0)
	var cfg string

	err := filepath.WalkDir(
		confPath,
		func(p string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}

			if filepath.Ext(d.Name()) != ".toml" {
				return nil
			}

			cfg = fmt.Sprintf("%s%s", confPath, d.Name())
			log.Infof("discovered cfg: '%s'", cfg)

			cfgFiles = append(cfgFiles, cfg)

			return nil
		})

	if err != nil {
		return cfgFiles, err
	}

	return cfgFiles, nil
}
