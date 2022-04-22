NAME=ALL
PUSH=false
DEPLOY=true

install:
	cd ./cmd && ./build.sh $(NAME) $(PUSH) $(DEPLOY)

