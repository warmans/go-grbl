package grbl

import "testing"

func TestIsRealtimeCommand(t *testing.T) {
	type args struct {
		cmd []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "! is realtime",
			args: args{cmd: []byte("!")},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRealtimeCommand(tt.args.cmd); got != tt.want {
				t.Errorf("IsRealtimeCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
