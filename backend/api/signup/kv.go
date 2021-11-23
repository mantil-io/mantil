package signup

import (
	"log"

	"github.com/mantil-io/mantil.go"
)

type kv struct {
	r *kvOps
	a *kvOps
	w *kvOps
}

func (k *kv) Registrations() *kvOps {
	if k.r == nil {
		k.r = newKvOps(registrationPartition)
	}
	return k.r
}

func (k *kv) Activations() *kvOps {
	if k.a == nil {
		k.a = newKvOps(activationPartition)
	}
	return k.a
}

func (k *kv) Workspaces() *kvOps {
	if k.w == nil {
		k.w = newKvOps(workspacePartition)
	}
	return k.w
}

func newKvOps(partition string) *kvOps {
	kv, err := mantil.NewKV(partition)
	if err != nil {
		log.Printf("mantil.NewKV failed: %s", err)
	}
	return &kvOps{connectError: err, kv: kv}
}

type kvOps struct {
	connectError error
	kv           *mantil.KV
}

func (k *kvOps) Put(id string, rec interface{}) error {
	if k.connectError != nil {
		return internalServerError
	}
	if err := k.kv.Put(id, rec); err != nil {
		log.Printf("kv.Put failed: %s", err)
		return internalServerError
	}
	return nil
}

func (k *kvOps) Get(id string, rec interface{}) error {
	if k.connectError != nil {
		return internalServerError
	}
	if err := k.kv.Get(id, &rec); err != nil {
		log.Printf("kv.Get failed: %s", err)
		return err
	}
	return nil
}
