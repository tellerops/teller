package providers

import (
	"github.com/99designs/keyring"
	"github.com/golang/mock/gomock"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
	"testing"
)

func TestMacKeychainProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockKeychainClient(ctrl)
	out := keyring.Item{Key: "MG_KEY", Data: []byte("shazam")}
	out2 := keyring.Item{Key: "SMTP_PASS", Data: []byte("mailman")}
	outKeys := []string{"SMTP_PASS", "MG_KEY"}

	client.EXPECT().Get(gomock.Eq("MG_KEY")).Return(out, nil).AnyTimes()
	client.EXPECT().Get(gomock.Eq("SMTP_PASS")).Return(out2, nil).AnyTimes()
	client.EXPECT().Keys().Return(outKeys, nil).AnyTimes()

	s := MacKeychain{
		client: client,
		logger: GetTestLogger(),
	}
	AssertProvider(t, &s, true)
}
