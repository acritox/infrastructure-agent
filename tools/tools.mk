.PHONY: ec2
ec2:
	go build -o tools/spin-ec2/bin tools/spin-ec2/*.go && tools/spin-ec2/bin