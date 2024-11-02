MAKEFLAGS+="-j 2"

init:
	go mod tidy
	@pnpm install

dev-go:
	@air

dev-solid:
	@pnpm run dev

dev: dev-solid dev-go

prod-go:
	go build .

prod-solid:
	@pnpm run build

prod: prod-go prod-solid

clean:
	rm domains
