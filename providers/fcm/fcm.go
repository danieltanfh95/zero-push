package fcm

import (
	"encoding/base64"
	"github.com/berty/zero-push"
	zpErrors "github.com/berty/zero-push/errors"
	"github.com/berty/zero-push/proto"
	"github.com/pkg/errors"
	"strings"

	"github.com/NaySoftware/go-fcm"
)

type FCMDispatcher struct {
	client      *fcm.FcmClient
	appID       string
	jsonDataKey string
}

func NewFCMDispatcher(appIDApiKey string, jsonDataKey string) (push.Dispatcher, error) {
	splitResult := strings.SplitN(appIDApiKey, ":", 2)
	if len(splitResult) != 2 {
		return nil, zpErrors.ErrPushInvalidServerConfig
	}

	appID := splitResult[0]
	apiKey := splitResult[1]

	client := fcm.NewFcmClient(apiKey)

	dispatcher := &FCMDispatcher{
		client:      client,
		appID:       appID,
		jsonDataKey: jsonDataKey,
	}

	return dispatcher, nil
}

func (d *FCMDispatcher) CanDispatch(pushDestination *proto.PushDestination) bool {
	if pushDestination.PushType != proto.DevicePushType_FCM {
		return false
	}

	fcmIdentifier := &proto.PushNativeIdentifier{}
	if err := fcmIdentifier.Unmarshal(pushDestination.PushId); err != nil {
		return false
	}

	if d.appID != fcmIdentifier.PackageID {
		return false
	}

	return true
}

func (d *FCMDispatcher) Dispatch(pushData *proto.PushData, pushDestination *proto.PushDestination) error {
	fcmIdentifier := &proto.PushNativeIdentifier{}
	if err := fcmIdentifier.Unmarshal(pushDestination.PushId); err != nil {
		return errors.Wrap(err, zpErrors.ErrPushUnknownDestination.Error())
	}

	payload := map[string]string{
		d.jsonDataKey: base64.StdEncoding.EncodeToString(pushData.Envelope),
	}

	deviceToken := fcmIdentifier.DeviceToken
	if _, err := d.client.NewFcmMsgTo(deviceToken, payload).Send(); err != nil {
		return errors.Wrap(err, zpErrors.ErrPushProvider.Error())
	}

	return nil
}

var _ push.Dispatcher = &FCMDispatcher{}
