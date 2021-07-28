package stream

import (
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

func Subscribe(topic string, handler func(nm *nats.Msg)) error {
	url := "connect.mantil.team"
	nc, err := nats.Connect(url, natsAuth())
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	sub, err := nc.Subscribe(topic, func(nm *nats.Msg) {
		if len(nm.Data) == 0 {
			wg.Done()
			return
		}
		handler(nm)
	})
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()
	wg.Wait()
	return nil
}

func natsAuth() nats.Option {
	nkeySeed := "SUADEJU2KOHEGHFLBXJS4QF75E2II3PU63I3GCK4OBJLINOC7LDVEOX42A"
	nkeyUser := "UDQPHZBVNZCJSM5JXUDICFALEQ7Y5KPPAF7KGHTH77OGG7COQOJEYBZ7"

	opt := nats.Nkey(nkeyUser, func(nonce []byte) ([]byte, error) {
		user, err := nkeys.FromSeed([]byte(nkeySeed))
		if err != nil {
			return nil, err
		}
		return user.Sign(nonce)
	})
	return opt
}
