package providers

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
	spb "go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestEtcd(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockEtcdClient(ctrl)
	path := "settings/prod/billing-svc"
	pathmap := "settings/prod/billing-svc/all"

	kv1 := &spb.KeyValue{
		Key:   []byte("settings/prod/billing-svc"),
		Value: []byte("shazam"),
	}
	kv2 := &spb.KeyValue{
		Key:   []byte("settings/prod/billing-svc"),
		Value: []byte("mailman"),
	}
	out := clientv3.GetResponse{
		Kvs: []*spb.KeyValue{kv1},
	}
	outmap := clientv3.GetResponse{
		Kvs: []*spb.KeyValue{kv2, kv1},
	}
	client.EXPECT().Get(gomock.Any(), gomock.Eq(path)).Return(&out, nil).AnyTimes()
	client.EXPECT().Get(gomock.Any(), gomock.Eq(pathmap), gomock.Any()).Return(&outmap, nil).AnyTimes()
	s := Etcd{
		client: client,
		logger: GetTestLogger(),
	}
	AssertProvider(t, &s, true)
}

func TestEtcdFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockEtcdClient(ctrl)
	client.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).AnyTimes()
	s := Etcd{
		client: client,
		logger: GetTestLogger(),
	}
	_, err := s.Get(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)
}
