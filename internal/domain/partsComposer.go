package domain

import (
	"context"
	"github.com/satmaelstorm/filup/internal/domain/dto"
	"github.com/satmaelstorm/filup/internal/domain/port"
)

type PartsComposer struct {
	storage port.PartsComposer
	cfg     port.UploaderConfig
	in      chan dto.UploaderStartResult
	logger  port.CriticalLogger
}

func ProvidePartsComposer(
	ctx port.ContextProvider,
	storage port.PartsComposer,
	cfg port.UploaderConfig,
	logger port.CriticalLogger,
) *PartsComposer {
	pc := new(PartsComposer)
	pc.storage = storage
	pc.cfg = cfg
	pc.in = make(chan dto.UploaderStartResult, cfg.GetComposerWorkers()*2)
	pc.logger = logger

	pc.runWorkers(ctx.Ctx())

	return pc
}

func (pc *PartsComposer) Run(metaInfo dto.UploaderStartResult) {
	pc.in <- metaInfo
}

func (pc *PartsComposer) runWorkers(ctx context.Context) {
	for i := 0; i < pc.cfg.GetComposerWorkers(); i++ {
		go pc.worker(ctx, pc.in)
	}
}

func (pc *PartsComposer) worker(ctx context.Context, in <-chan dto.UploaderStartResult) {
	for {
		select {
		case <-ctx.Done():
			return
		case metaInfo := <-in:
			pc.process(metaInfo)
		}
	}
}

func (pc *PartsComposer) process(metaInfo dto.UploaderStartResult) {
	partsNames := make([]string, 0, len(metaInfo.GetChunks()))
	for pn, _ := range metaInfo.GetChunks() {
		partsNames = append(partsNames, pn)
	}
	_, err := pc.storage.ComposeFileParts(
		metaInfo.GetUUID(),
		partsNames,
		metaInfo.GetUserTags(),
	)
	if err != nil {
		pc.logger.Critical().Println(err)
	}
	//TODO callbacks
	//TODO add cleaner
}
