package providers

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/v-braun/enpass-cli/pkg/enpass"

	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestEnpass(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockEnpassClient(ctrl)
	card := mock_providers.NewMockEnpassCard(ctrl)
	path := "settings/prod/billing-svc"
	// pathmap := "settings/prod/billing-svc/all"
	out := enpass.Card{
		RawValue: "shazam",
	}
	outList := []enpass.Card{
		out,
	}
	client.EXPECT().GetEntry(path, gomock.Any(), true).Return(out, nil).AnyTimes()
	client.EXPECT().GetEntries(path, gomock.Any()).Return(outList, nil).AnyTimes()
	card.EXPECT().Decrypt().Return("shazam", nil).AnyTimes()
	s := Enpass{
		client: client,
		logger: GetTestLogger(),
	}
	AssertProvider(t, &s, true)
}

// func TestEtcdFailures(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	// Assert that Bar() is invoked.
// 	defer ctrl.Finish()
// 	client := mock_providers.NewMockEtcdClient(ctrl)

// 	assert.NotNil(t, err)
// }
