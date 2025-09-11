module github.com/gaboose/afero-guestfs

go 1.23.0

toolchain go1.24.7

replace libguestfs.org/guestfs => ./libguestfs.org/guestfs

require (
	github.com/spf13/afero v1.15.0
	github.com/stretchr/testify v1.11.1
	libguestfs.org/guestfs v1.56.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
