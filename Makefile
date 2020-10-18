clean:
	rm -f .data/cache.db
	rm -f ./kinopub-core

build:
	go build

release:
	GOOS=linux GOARCH=arm GOARM=7 go build

pi-deploy:
	ssh pi@raspberrypi.local 'mkdir -p /home/pi/Projects/kinohub-core'
	scp kinohub-core pi@raspberrypi.local:/home/pi/Projects/kinohub-core
	ssh pi@raspberrypi.local 'sudo service kinohub stop'
	ssh pi@raspberrypi.local 'sudo cp /home/pi/Projects/kinohub-core/kinohub-core /usr/bin/'
	ssh pi@raspberrypi.local 'sudo service kinohub start'

pi-init:
	ssh pi@raspberrypi.local 'sudo mkdir -p /var/lib/kinohub/data'
	ssh pi@raspberrypi.local 'sudo chmod u+w /var/lib/kinohub'