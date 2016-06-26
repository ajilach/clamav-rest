SERVICE=clamrest
VERSION := dev

run-container: .clamav build-container
	-@docker rm -f $(SERVICE)
	@docker run -d -p 9000:9000 -e PORT=9000 --name $(SERVICE) --link clamd:clamd $(SERVICE):$(VERSION)

build-container:
	docker build -t $(SERVICE):$(VERSION) .

run-slug: .clamav build-slug
	-@docker rm -f $(SERVICE)  
	@docker run -d -v target/app:/app -p 9000:9000 -e PORT=9000 --name $(SERVICE) --link clamd:clamd flynn/slugrunner start web 
	@echo "Clamrest listening on port 9000"

build-slug:
	-@rm -rf target
	-@mkdir target
	@tar cf - . | docker run --rm -i -a stdin -a stdout -a stderr flynn/slugbuilder -> target/slug.tgz

test: 
	@rm -rf tests/pyenv
	@virtualenv tests/pyenv
	@. tests/pyenv/bin/activate; pip install -r tests/requirements.txt
	@cd tests; . pyenv/bin/activate; behave

.clamav:
	@echo "Starting clamav docker image"
	-@docker rm -f clamd
	@docker run -d -p 3310:3310 --name clamd dinkel/clamavd 
	@echo "Waiting for clamd to respond"
	@sleep 10

.restapi:
	docker 
