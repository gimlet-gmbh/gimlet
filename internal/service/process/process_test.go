package process

import "testing"

func TestLocalManager(t *testing.T) {
	var p Manager
	p = NewLocalBinaryManager("test", "/", "/", []string{}, []string{})
	pid, err := p.Start()
	if err == nil {
		t.Fatalf("process should not report starting without actual dummy details")
	}
	if pid != -1 {
		t.Fatalf("process should not report non -1 pid")
	}
}
