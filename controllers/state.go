package controllers

import (
	"encoding/json"
	"github.com/mmlt/operator-addons/internal/cluster"
	"k8s.io/apimachinery/pkg/api/errors"
)

// State is the current state of the target cluster.
type state struct {
	// Sources maps a CR source name to the current state.
	Sources map[string]sourceState
}
type sourceState struct {
	// RepoSHA is the SHA of the last applied repository.
	RepoSHA string
	// ActionHash is the hash of the last applied action.
	ActionHash uint64
}

// FieldName in cluster state ConfigMap
const fieldName = "op-addons"

// GetState reads the state from the target cluster.
// GetState returns an empty state when no state is available in the target cluster.
func getState(target *cluster.Cluster) (*state, error) {
	emptyState := state{Sources: map[string]sourceState{}}

	m, err := target.GetState()
	if err != nil {
		if isCode(err, 404) {
			// ConfigMap not found.
			return &emptyState, nil
		}
		return nil, err
	}

	b, ok := m[fieldName]
	if !ok {
		return &emptyState, err
	}

	var state state
	err = json.Unmarshal([]byte(b), &state)
	if err != nil {
		return &emptyState, err
	}

	return &state, nil
}

// PutState writes the state to the target cluster.
func putState(target *cluster.Cluster, st *state) error {
	b, err := json.Marshal(st)
	if err != nil {
		return err
	}

	return target.PutState(map[string]string{fieldName: string(b)})
}

// IsCode answers true when err is an StatusError with HTTP result code.
func isCode(err error, code int32) bool {
	switch t := err.(type) {
	case *errors.StatusError:
		return t.ErrStatus.Code == code
	default:
		return false
	}
}
