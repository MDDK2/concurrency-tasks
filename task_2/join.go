package task_2

import (
	"context"
	"reflect"
)

func join(ctx context.Context, channels ...chan interface{}) chan []interface{} {
	joinedChan := make(chan []interface{})
	if len(channels) == 0 {
		close(joinedChan)
		return joinedChan
	}
	go func() {
		const Done = 0
		for {
			cases := make([]reflect.SelectCase, 1, len(channels)+1)
			cases[Done] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(ctx.Done()),
			}
			for _, ch := range channels {
				cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)})
			}

			var (
				completed = 0
				joined    = make([]interface{}, len(channels))
			)

			for {
				if completed == len(channels) {
					select {
					case <-ctx.Done():
						return
					case joinedChan <- joined:
					}
					break
				}
				chosen, value, ok := reflect.Select(cases)
				if !ok {
					return
				}
				switch chosen {
				case Done:
					return
				default:
					joined[chosen-1] = value.Interface()
					cases[chosen].Chan = reflect.ValueOf(nil)
					completed++
				}
			}
		}
	}()
	return joinedChan
}
