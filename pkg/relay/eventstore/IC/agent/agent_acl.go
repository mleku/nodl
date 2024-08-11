package agent

import "time"

func (b *Backend) AddUser(pubKey string, perm bool) (err error) {
	methodName := "add_user"
	args := []any{pubKey, perm, time.Now().UnixNano()}
	var result *string
	if err = b.Agent.Call(b.CanisterID, methodName, args, []any{&result}); chk.E(err) {
		return
	} else if result != nil {
		err = log.E.Err("failed to add user")
		return
	}
	return nil
}

func (b *Backend) RemoveUser(pubKey string) (err error) {
	methodName := "remove_user"
	args := []any{pubKey, time.Now().UnixNano()}
	var result *string
	if err = b.Agent.Call(b.CanisterID, methodName, args, []any{&result}); chk.E(err) {
		return
	} else if result != nil {
		err = log.E.Err("failed to remove user")
		return
	}
	return nil
}

func (b *Backend) GetPermission() (result string, err error) {
	methodName := "get_permission"
	args := []any{time.Now().UnixNano()}
	if err = b.Agent.Query(b.CanisterID, methodName, args, []any{&result}); chk.E(err) {
		return
	}
	return
}
