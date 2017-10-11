VERSION      = $(shell git show -q --format=%h)
REPOSITORY   = registry.bukalapak.io/bukalapak/protor

tmp:
	mkdir tmp

clean:
	rm -r tmp

test:
	govendor test -v -cover +local,^program

dep:
	govendor fetch -v +outside

aggregator-start:
	docker run -d --name="protor" --network=host run rolandhawk/prometheus-aggregator

aggregator-stop:
	docker rm -f protor
