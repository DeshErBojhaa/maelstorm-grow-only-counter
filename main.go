package main

import (
	"context"
	"encoding/json"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
)

type AddMsg struct {
	Type  string `json:"type"`
	Delta int    `json:"delta"`
}

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(n)
	key := n.ID()

	n.Handle("latest", func(msg maelstrom.Message) error {
		val, err := kv.ReadInt(context.Background(), key)
		if err != nil && maelstrom.ErrorCode(err) != maelstrom.KeyDoesNotExist {
			val = 0
		}

		return n.Reply(msg, map[string]int{"value": val})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		maxVal, err := kv.ReadInt(context.Background(), key)
		if err != nil && maelstrom.ErrorCode(err) != maelstrom.KeyDoesNotExist {
			panic(err)
			return err
		}
		for _, k := range n.NodeIDs() {
			if k == n.ID() {
				continue
			}
			resp, err := n.SyncRPC(context.Background(), k, map[string]string{"type": "latest"})
			if err != nil {
				return err
			}
			body := make(map[string]int)
			if err := json.Unmarshal(resp.Body, &body); err != nil {
				return err
			}
			maxVal += body["value"]
		}

		return n.Reply(msg, map[string]any{
			"type":  "read_ok",
			"value": maxVal,
		})
	})

	n.Handle("add", func(msg maelstrom.Message) error {
		var body AddMsg
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		val, err := kv.ReadInt(context.Background(), key)
		if err != nil && maelstrom.ErrorCode(err) != maelstrom.KeyDoesNotExist {
			panic(err)
			return err
		}

		if err := kv.Write(context.Background(), key, val+body.Delta); err != nil {
			return err
		}

		return n.Reply(msg, map[string]any{"type": "add_ok"})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
