package shipping

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYanwenPublishedChannelNormalizesAndValidatesEnvironment(t *testing.T) {
	productionChannel := YanwenPublishedChannel{ProductCode: "481", DisplayName: "燕文生产普货", VolumetricDivisor: 8000}
	require.NoError(t, productionChannel.Validate())
	require.Equal(t, YanwenPublishedChannelEnvironmentProduction, productionChannel.Environment)

	fatChannel := YanwenPublishedChannel{Environment: " FAT ", ProductCode: "481", DisplayName: "燕文 FAT 普货", VolumetricDivisor: 8000}
	require.NoError(t, fatChannel.Validate())
	require.Equal(t, YanwenPublishedChannelEnvironmentFAT, fatChannel.Environment)

	invalidChannel := YanwenPublishedChannel{Environment: "staging", ProductCode: "482", DisplayName: "无效环境渠道", VolumetricDivisor: 8000}
	require.ErrorContains(t, invalidChannel.Validate(), "environment must be fat or production")
}
