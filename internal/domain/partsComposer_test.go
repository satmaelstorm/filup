package domain

import (
	"github.com/satmaelstorm/filup/internal/domain/dto"
	"github.com/stretchr/testify/suite"
	"testing"
)

type suitePartsComposer struct {
	suite.Suite
	pc *PartsComposer
}

func TestPartsComposer(t *testing.T) {
	suite.Run(t, new(suitePartsComposer))
}

func (s *suitePartsComposer) SetupSuite() {
}

func (s *suitePartsComposer) TestGetChunksSlice() {
	pc := new(PartsComposer)
	names := pc.getChunksSlice(dto.UploaderStartResult{
		Uuid:     "",
		Size:     0,
		UserTags: nil,
		Chunks: map[string]dto.UploaderChunk{
			"part2": dto.NewUploaderChunk("part2", 100, 50),
			"part0": dto.NewUploaderChunk("part0", 10, 0),
			"part3": dto.NewUploaderChunk("part3", 100, 150),
			"part1": dto.NewUploaderChunk("part1", 40, 10),
		},
	})
	s.Require().Equal(4, len(names))
	s.Equal("part0", names[0])
	s.Equal("part1", names[1])
	s.Equal("part2", names[2])
	s.Equal("part3", names[3])
}
