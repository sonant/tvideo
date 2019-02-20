build:
	go build -o main 
	docker build -t tvideo .
run:
	docker run -d --rm --name tvideo -v /etc/ssl:/etc/ssl:ro -v /opt/videos:/opt/videos -p 80:8080 tvideo
