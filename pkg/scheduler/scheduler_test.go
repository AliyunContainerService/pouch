package scheduler

import (
	"context"
	"reflect"
	"sync"
	"testing"
)

func TestNewLRUScheduler(t *testing.T) {
	type args struct {
		pool []Factory
	}
	tests := []struct {
		name    string
		args    args
		want    Scheduler
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLRUScheduler(tt.args.pool)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLRUScheduler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLRUScheduler() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testFactory struct {
	mux sync.Mutex

	data int
}

func (tf *testFactory) Consume(v int) error {
	tf.mux.Lock()
	defer tf.mux.Unlock()

	tf.data -= v
	return nil
}

func (tf *testFactory) Produce(v int) {
	tf.mux.Lock()
	defer tf.mux.Unlock()

	tf.data += v
}

func (tf *testFactory) Value() int {
	tf.mux.Lock()
	defer tf.mux.Unlock()

	return tf.data
}

func newTestFactoryPool(ls []int) []Factory {
	pool := []Factory{}
	for _, d := range ls {
		pool = append(pool, &testFactory{
			data: d,
		})
	}

	return pool
}

func TestLRUScheduler_Schedule(t *testing.T) {
	type fields struct {
		pool []Factory
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Factory
		wantErr bool
	}{
		{name: "test1", fields: fields{pool: newTestFactoryPool([]int{})}, args: args{ctx: context.Background()}, want: nil, wantErr: true},
		{name: "test1", fields: fields{pool: newTestFactoryPool([]int{1, 2, 3})}, args: args{ctx: context.Background()}, want: &testFactory{data: 3}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lru := &LRUScheduler{
				pool: tt.fields.pool,
			}
			got, err := lru.Schedule(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("LRUScheduler.Schedule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (tt.want == nil && got != nil) || (tt.want != nil && got == nil) {
				t.Errorf("LRUScheduler.Schedule() = %v, want %v", got, tt.want)
			} else if tt.want == nil && got == nil {
				return
			}

			gotTestFactory, ok := got.(*testFactory)
			if !ok {
				t.Errorf("LRUScheduler.Schedule() return type is wrong")
			}

			wantFactory, ok := tt.want.(*testFactory)
			if !ok {
				t.Errorf("tt.want not *testFactory")
			}

			if gotTestFactory.Value() != wantFactory.Value() {
				t.Errorf("LRUScheduler.Schedule() = %v, want %v", gotTestFactory, wantFactory)
			}
		})
	}
}
