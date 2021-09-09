bench:
	go test -run=NewUnmarshal -bench=NewUnmarshalString -benchmem -memprofile=mem.out -cpuprofile=cpu.out
test:
	go test -run=NewUnmarshal