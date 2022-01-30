package domain

import (
	"context"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/satmaelstorm/filup/internal/domain/dto"
	"github.com/satmaelstorm/filup/internal/domain/port"
	"net/url"
	"strconv"
	"strings"
)

type PartsComposer struct {
	storage port.PartsComposer
	cleaner port.StorageCleaner
	cfg     port.UploaderConfig
	in      chan dto.UploaderStartResult
	logger  port.Logger
	poster  port.Poster
	ctx     context.Context
}

func ProvidePartsComposer(
	ctx port.ContextProvider,
	storage port.PartsComposer,
	cleaner port.StorageCleaner,
	cfg port.UploaderConfig,
	logger port.Logger,
	poster port.Poster,
) *PartsComposer {
	pc := new(PartsComposer)
	pc.storage = storage
	pc.cfg = cfg
	pc.in = make(chan dto.UploaderStartResult, cfg.GetComposerWorkers()*2)
	pc.logger = logger
	pc.poster = poster
	pc.ctx = ctx.Ctx()
	pc.cleaner = cleaner

	pc.runWorkers(pc.ctx)

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
		pc.logger.Critical().Println(errors.Wrap(err, "PartsComposer.process()"))
	}
	if callbackAfter := pc.cfg.GetCallbackAfter(); callbackAfter != nil {
		pc.processCallbackAfter(callbackAfter, metaInfo) //TODO make async?
	}
	err = pc.cleaner.RemoveMeta(MetaFileName(metaInfo.GetUUID()))
	if err != nil {
		pc.logger.Error().Println(errors.Wrap(err, "PartsComposer.process()"))
	}
	err = pc.cleaner.RemoveParts(partsNames)
	if err != nil {
		pc.logger.Error().Println(errors.Wrap(err, "PartsComposer.process()"))
	}
}

func (pc *PartsComposer) processCallbackAfter(callbackAfter *url.URL, metaInfo dto.UploaderStartResult) {
	body, err := jsoniter.Marshal(metaInfo)
	if err != nil {
		pc.logger.Critical().Println(errors.Wrap(err, "PartsComposer.processCallbackAfter()"))
		return
	}
	retires := 0
	totalRetires := pc.cfg.GetHttpRetries()
	var allErrors []string
	for retires < totalRetires {
		_, code, err := pc.poster.Post(pc.ctx, *callbackAfter, pc.cfg.GetHttpTimeout(), body)
		if err == nil && (code >= 200 && code <= 299) {
			return
		}
		retires++
		if err != nil {
			e := errors.Wrap(err, "PartsComposer.processCallbackAfter().poster.error")
			pc.logger.Error().Println(e)
			allErrors = append(allErrors, e.Error())
			continue
		}
		if code < 200 || code > 299 {
			e := errors.Wrap(err, "PartsComposer.processCallbackAfter().poster.code_"+strconv.Itoa(code))
			pc.logger.Error().Println(e)
			allErrors = append(allErrors, e.Error())
			continue
		}
	}
	pc.logger.Critical().Println("CallbackAfter " + callbackAfter.String() +
		" Error after " + strconv.Itoa(totalRetires) + " with body " + string(body) +
		" with errors [" + strings.Join(allErrors, ",") + "]")
}
