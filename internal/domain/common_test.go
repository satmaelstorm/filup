package domain

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type commonTestSuite struct {
	suite.Suite
}

func TestCommon(t *testing.T) {
	suite.Run(t, new(commonTestSuite))
}

func (s *commonTestSuite) TestIsValidUuid() {
	s.True(IsCorrectUuid("870915da-76bb-11ec-8686-e4e7494803df"))
	s.True(IsCorrectUuid("870915da76bb11ec8686e4e7494803df"))
	s.True(IsCorrectUuid(ProvideUuidProvider().NewUuid()))
	s.False(IsCorrectUuid("870915da-76bb-11ec-8686_e4e7494803df"))
}

func (s *commonTestSuite) TestExtractUuidFromPartName() {
	var uuid string
	var err error
	uuid, err = ExtractUuidFromPartName("870915da-76bb-11ec-8686-e4e7494803df_part_0")
	s.Require().Nil(err)
	s.Equal("870915da-76bb-11ec-8686-e4e7494803df", uuid)

	uuid, err = ExtractUuidFromPartName("870915da-76bb-11ec-8686-e4e7494803dfpart_0")
	s.NotNil(err)

	myUuid := ProvideUuidProvider().NewUuid()
	partName := ChunkFileName(myUuid, 0)
	uuid, err = ExtractUuidFromPartName(partName)
	s.Require().Nil(err)
	s.Equal(myUuid, uuid)
}
