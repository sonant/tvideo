build:
	go build -o main 
	docker build -t tvideo .
run:
	go build -o main 
	docker build -t tvideo .
	docker kill tvideo
	docker run -d --rm --name tvideo -v /etc/ssl:/etc/ssl:ro -v videos:/opt/videos -p 8080:8080 tvideo
	docker logs -f tvideo