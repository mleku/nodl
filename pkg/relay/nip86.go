package relay

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"
	"strings"

	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/tag"
	"git.replicatr.dev/pkg/codec/timestamp"
	"git.replicatr.dev/pkg/util/context"
	"github.com/nbd-wtf/go-nostr/nip86"
)

type RelayManagementAPI struct {
	RejectAPICall []func(c context.T, mp nip86.MethodParams) (reject bool, msg string)

	BanPubKey                   func(c context.T, pubkey string, reason string) error
	ListBannedPubKeys           func(c context.T) ([]nip86.PubKeyReason, error)
	AllowPubKey                 func(c context.T, pubkey string, reason string) error
	ListAllowedPubKeys          func(c context.T) ([]nip86.PubKeyReason, error)
	ListEventsNeedingModeration func(c context.T) ([]nip86.IDReason, error)
	AllowEvent                  func(c context.T, id string, reason string) error
	BanEvent                    func(c context.T, id string, reason string) error
	ListBannedEvents            func(c context.T) ([]nip86.IDReason, error)
	ChangeRelayName             func(c context.T, name string) error
	ChangeRelayDescription      func(c context.T, desc string) error
	ChangeRelayIcon             func(c context.T, icon string) error
	AllowKind                   func(c context.T, kind int) error
	DisallowKind                func(c context.T, kind int) error
	ListAllowedKinds            func(c context.T) ([]int, error)
	BlockIP                     func(c context.T, ip net.IP, reason string) error
	UnblockIP                   func(c context.T, ip net.IP, reason string) error
	ListBlockedIPs              func(c context.T) ([]nip86.IPReason, error)
}

func (rl *Relay) HandleNIP86(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/nostr+json+rpc")

	var (
		resp        nip86.Response
		ctx         = r.Context()
		req         nip86.Request
		mp          nip86.MethodParams
		evt         event.T
		payloadHash [32]byte
	)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Error = "empty request"
		goto respond
	}
	payloadHash = sha256.Sum256(payload)

	{
		auth := r.Header.Get("Authorization")
		spl := strings.Split(auth, "Nostr ")
		if len(spl) != 2 {
			resp.Error = "missing auth"
			goto respond
		}
		if evtj, err := base64.StdEncoding.DecodeString(spl[1]); err != nil {
			resp.Error = "invalid base64 auth"
			goto respond
		} else if err := json.Unmarshal(evtj, &evt); err != nil {
			resp.Error = "invalid auth event json"
			goto respond
		} else if ok, _ := evt.CheckSignature(); !ok {
			resp.Error = "invalid auth event"
			goto respond
		} else if uTag := evt.Tags.GetFirst(tag.New("u", "")); uTag == nil || getServiceBaseURL(r) != S(uTag.Field[1]) {
			resp.Error = "invalid 'u' tag"
			goto respond
		} else if pht := evt.Tags.GetFirst(tag.New("payload", hex.EncodeToString(payloadHash[:]))); pht == nil {
			resp.Error = "invalid auth event payload hash"
			goto respond
		} else if evt.CreatedAt.Int() < timestamp.Now().Int()-30 {
			resp.Error = "auth event is too old"
			goto respond
		}
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		resp.Error = "invalid json body"
		goto respond
	}

	mp, err = nip86.DecodeRequest(req)
	if err != nil {
		resp.Error = fmt.Sprintf("invalid params: %s", err)
		goto respond
	}

	ctx = context.Value(ctx, nip86HeaderAuthKey, evt.PubKey)
	for _, rac := range rl.ManagementAPI.RejectAPICall {
		if reject, msg := rac(ctx, mp); reject {
			resp.Error = msg
			goto respond
		}
	}

	if _, ok := mp.(nip86.SupportedMethods); ok {
		mat := reflect.TypeOf(rl.ManagementAPI)
		mav := reflect.ValueOf(rl.ManagementAPI)

		methods := make([]string, 0, mat.NumField())
		for i := 0; i < mat.NumField(); i++ {
			field := mat.Field(i)

			// danger: this assumes the struct fields are appropriately named
			methodName := strings.ToLower(field.Name)

			// assign this only if the function was defined
			if mav.Field(i).Interface() != nil {
				methods[i] = methodName
			}
		}
		resp.Result = methods
	} else {
		switch thing := mp.(type) {
		case nip86.BanPubKey:
			if rl.ManagementAPI.BanPubKey == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if err := rl.ManagementAPI.BanPubKey(ctx, thing.PubKey, thing.Reason); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = true
			}
		case nip86.ListBannedPubKeys:
			if rl.ManagementAPI.ListBannedPubKeys == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if result, err := rl.ManagementAPI.ListBannedPubKeys(ctx); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = result
			}
		case nip86.AllowPubKey:
			if rl.ManagementAPI.AllowPubKey == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if err := rl.ManagementAPI.AllowPubKey(ctx, thing.PubKey, thing.Reason); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = true
			}
		case nip86.ListAllowedPubKeys:
			if rl.ManagementAPI.ListAllowedPubKeys == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if result, err := rl.ManagementAPI.ListAllowedPubKeys(ctx); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = result
			}
		case nip86.BanEvent:
			if rl.ManagementAPI.BanEvent == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if err := rl.ManagementAPI.BanEvent(ctx, thing.ID, thing.Reason); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = true
			}
		case nip86.AllowEvent:
			if rl.ManagementAPI.AllowEvent == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if err := rl.ManagementAPI.AllowEvent(ctx, thing.ID, thing.Reason); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = true
			}
		case nip86.ListEventsNeedingModeration:
			if rl.ManagementAPI.ListEventsNeedingModeration == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if result, err := rl.ManagementAPI.ListEventsNeedingModeration(ctx); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = result
			}
		case nip86.ListBannedEvents:
			if rl.ManagementAPI.ListBannedEvents == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if result, err := rl.ManagementAPI.ListEventsNeedingModeration(ctx); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = result
			}
		case nip86.ChangeRelayName:
			if rl.ManagementAPI.ChangeRelayName == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if err := rl.ManagementAPI.ChangeRelayName(ctx, thing.Name); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = true
			}
		case nip86.ChangeRelayDescription:
			if rl.ManagementAPI.ChangeRelayDescription == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if err := rl.ManagementAPI.ChangeRelayDescription(ctx, thing.Description); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = true
			}
		case nip86.ChangeRelayIcon:
			if rl.ManagementAPI.ChangeRelayIcon == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if err := rl.ManagementAPI.ChangeRelayIcon(ctx, thing.IconURL); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = true
			}
		case nip86.AllowKind:
			if rl.ManagementAPI.AllowKind == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if err := rl.ManagementAPI.AllowKind(ctx, thing.Kind); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = true
			}
		case nip86.DisallowKind:
			if rl.ManagementAPI.DisallowKind == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if err := rl.ManagementAPI.DisallowKind(ctx, thing.Kind); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = true
			}
		case nip86.ListAllowedKinds:
			if rl.ManagementAPI.ListAllowedKinds == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if result, err := rl.ManagementAPI.ListAllowedKinds(ctx); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = result
			}
		case nip86.BlockIP:
			if rl.ManagementAPI.BlockIP == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if err := rl.ManagementAPI.BlockIP(ctx, thing.IP, thing.Reason); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = true
			}
		case nip86.UnblockIP:
			if rl.ManagementAPI.UnblockIP == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if err := rl.ManagementAPI.UnblockIP(ctx, thing.IP, thing.Reason); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = true
			}
		case nip86.ListBlockedIPs:
			if rl.ManagementAPI.ListBlockedIPs == nil {
				resp.Error = fmt.Sprintf("method %s not supported", thing.MethodName())
			} else if result, err := rl.ManagementAPI.ListBlockedIPs(ctx); err != nil {
				resp.Error = err.Error()
			} else {
				resp.Result = result
			}
		default:
			resp.Error = fmt.Sprintf("method '%s' not known", mp.MethodName())
		}
	}

respond:
	json.NewEncoder(w).Encode(resp)
}
