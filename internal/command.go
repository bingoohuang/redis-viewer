package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bingoohuang/jj"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cast"
)

type errMsg struct {
	err error
}

type scanMsg struct {
	items []list.Item
}

func (m model) scanCmd() tea.Cmd {
	return func() tea.Msg {
		var (
			val   any
			err   error
			items []list.Item
		)

		ctx := context.Background()
		keyMessages := GetKeys(m.rdb, cast.ToUint64(m.offset*m.limit), m.searchValue, m.limit)
		for keyMessage := range keyMessages {
			if keyMessage.Err != nil {
				return errMsg{err: keyMessage.Err}
			}
			kt := m.rdb.Type(ctx, keyMessage.Key).Val()
			switch kt {
			case "string":
				val, err = m.rdb.Get(ctx, keyMessage.Key).Result()
			case "list":
				val, err = m.rdb.LRange(ctx, keyMessage.Key, 0, -1).Result()
			case "set":
				val, err = m.rdb.SMembers(ctx, keyMessage.Key).Result()
			case "zset":
				val, err = m.rdb.ZRange(ctx, keyMessage.Key, 0, -1).Result()
			case "hash":
				val, err = m.rdb.HGetAll(ctx, keyMessage.Key).Result()
			default:
				val = ""
				err = fmt.Errorf("unsupported type: %s", kt)
			}
			if err != nil {
				items = append(
					items,
					item{keyType: kt, key: keyMessage.Key, val: err.Error(), err: true},
				)
			} else {
				valBts, _ := json.MarshalIndent(val, "", "  ")
				valBts = jj.Pretty(jj.FreeInnerJSON(valBts))

				items = append(items, item{keyType: kt, key: keyMessage.Key, val: string(valBts)})
			}
		}

		return scanMsg{items: items}
	}
}

type okMsg struct {
	message string
}

type countMsg struct {
	count int
}

func (m model) okCmd(msg string) tea.Cmd {
	return func() tea.Msg {
		return okMsg{message: msg}
	}
}

func (m model) countCmd() tea.Cmd {
	return func() tea.Msg {
		count, err := CountKeys(m.rdb, m.searchValue)
		if err != nil {
			return errMsg{err: err}
		}

		return countMsg{count: count}
	}
}

type tickMsg struct {
	t string
}

func (m model) Close() error {
	if m.rdb != nil {
		err := m.rdb.Close()
		m.rdb = nil
		return err
	}

	return nil
}

func (m model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(_ time.Time) tea.Msg {
		return tickMsg{t: time.Now().Format("2006-01-02 15:04:05")}
	})
}
